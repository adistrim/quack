import { existsSync, mkdirSync, chmodSync, createWriteStream, unlinkSync } from "node:fs";
import { join, dirname } from "node:path";
import { homedir } from "node:os";
import { getPlatformTarget, getBinaryName, QuackBinaryError, checkBinaryStatus, type SupportedPlatform } from "./binary";

const GITHUB_RELEASE_BASE = "https://github.com/adistrim/quack/releases/download";

// Cache directory for auto-downloaded binaries (~/.cache/quack-search/)
function getCacheDir(): string {
  const home = homedir();
  if (process.platform === "win32") {
    return join(process.env.LOCALAPPDATA || join(home, "AppData", "Local"), "quack-search");
  }
  return join(process.env.XDG_CACHE_HOME || join(home, ".cache"), "quack-search");
}

// ---------- Internal Bootstrap (Prisma-style auto-download) ----------

let bootstrapPromise: Promise<string> | null = null;

/**
 * Internal: Ensures binary is ready, downloading automatically if needed.
 * This is called internally before any core execution.
 * Users should never need to call this directly.
 */
export async function ensureBinaryReady(): Promise<string> {
  // Return cached promise if bootstrap is in progress
  if (bootstrapPromise) {
    return bootstrapPromise;
  }

  bootstrapPromise = doBootstrap();
  try {
    return await bootstrapPromise;
  } catch (err) {
    // Reset on failure so next call can retry
    bootstrapPromise = null;
    throw err;
  }
}

async function doBootstrap(): Promise<string> {
  // check binary is already available 
  const status = checkBinaryStatus();
  if (status.found && status.path) {
    return status.path;
  }

  // download to cache directory
  const cacheDir = getCacheDir();
  const platform = getPlatformTarget();
  const binaryName = getBinaryName();
  const cachePath = join(cacheDir, binaryName);

  // Check if already in cache (may have been downloaded previously)
  if (existsSync(cachePath)) {
    try {
      const { accessSync, constants } = await import("node:fs");
      accessSync(cachePath, constants.X_OK);
      return cachePath;
    } catch {
      // Cache exists but not executable, will re-download
    }
  }

  // Download to cache
  try {
    const result = await downloadBinary({
      targetDir: cacheDir,
      version: "latest",
      platform,
      verbose: false,
    });
    return result.binaryPath;
  } catch (downloadErr) {
    // Provide clear, actionable error
    throw new QuackBinaryError(
      `Could not find or download quack binary for ${process.platform} ${process.arch}`,
      "BOOTSTRAP_FAILED",
      `Automatic download failed. To fix this:\n` +
        `  1. Check your internet connection\n` +
        `  2. Or install manually: npm install quack-search-${platform}\n` +
        `  3. Or set QUACK_BINARY_PATH to a pre-downloaded binary\n\n` +
        `Original error: ${downloadErr instanceof Error ? downloadErr.message : String(downloadErr)}`
    );
  }
}

/**
 * Resets the bootstrap state. Useful for testing or forcing re-download.
 * @internal
 */
export function resetBootstrap(): void {
  bootstrapPromise = null;
}

export interface DownloadOptions {
  /** Target directory for the binary. Defaults to package bin directory. */
  targetDir?: string;
  /** Specific version to download. Defaults to 'latest'. */
  version?: string;
  /** Override platform detection. */
  platform?: SupportedPlatform;
  /** Show progress output. */
  verbose?: boolean;
}

export interface DownloadResult {
  binaryPath: string;
  version: string;
  platform: SupportedPlatform;
}

/**
 * Downloads the quack binary for the current platform.
 * Use this for explicit binary setup in Docker, CI, or bundled environments.
 *
 * @example
 * ```ts
 * import { downloadBinary } from 'quack-search';
 *
 * // Download to default location
 * const result = await downloadBinary();
 * console.log(`Binary downloaded to: ${result.binaryPath}`);
 *
 * // Download to custom location
 * await downloadBinary({ targetDir: '/app/bin' });
 * ```
 */
export async function downloadBinary(
  options: DownloadOptions = {}
): Promise<DownloadResult> {
  const {
    targetDir = getDefaultBinDir(),
    version = "latest",
    platform = getPlatformTarget(),
    verbose = false,
  } = options;

  const binaryName = getBinaryName();
  const targetPath = join(targetDir, binaryName);

  // Ensure target directory exists
  if (!existsSync(targetDir)) {
    mkdirSync(targetDir, { recursive: true });
  }

  // Resolve version (could fetch latest tag from GitHub API)
  const resolvedVersion = version === "latest" ? await fetchLatestVersion() : version;

  // Build download URL
  const downloadUrl = buildDownloadUrl(platform, resolvedVersion);

  if (verbose) {
    console.log(`Downloading quack binary...`);
    console.log(`  Platform: ${platform}`);
    console.log(`  Version: ${resolvedVersion}`);
    console.log(`  URL: ${downloadUrl}`);
    console.log(`  Target: ${targetPath}`);
  }

  // Download the binary
  await downloadFile(downloadUrl, targetPath);

  // Make executable (non-Windows)
  if (process.platform !== "win32") {
    chmodSync(targetPath, 0o755);
  }

  if (verbose) {
    console.log(`Binary downloaded successfully!`);
  }

  return {
    binaryPath: targetPath,
    version: resolvedVersion,
    platform,
  };
}

async function fetchLatestVersion(): Promise<string> {
  try {
    const response = await fetch(
      "https://api.github.com/repos/adistrim/quack/releases/latest",
      {
        headers: {
          Accept: "application/vnd.github.v3+json",
          "User-Agent": "quack-search",
        },
      }
    );

    if (!response.ok) {
      throw new QuackBinaryError(
        `Failed to fetch latest version: GitHub API returned ${response.status}`,
        "VERSION_FETCH_FAILED",
        `Specify an explicit version instead, e.g.: downloadBinary({ version: 'v0.1.0' })`
      );
    }

    const data = (await response.json()) as { tag_name: string };
    if (!data.tag_name) {
      throw new QuackBinaryError(
        "Invalid response from GitHub API: missing tag_name",
        "VERSION_PARSE_FAILED",
        `Specify an explicit version instead, e.g.: downloadBinary({ version: 'v0.1.0' })`
      );
    }
    return data.tag_name;
  } catch (err) {
    if (err instanceof QuackBinaryError) {
      throw err;
    }
    throw new QuackBinaryError(
      "Failed to fetch latest version from GitHub",
      "VERSION_FETCH_FAILED",
      `Network error or GitHub unavailable. Specify an explicit version, e.g.: downloadBinary({ version: 'v0.1.0' })`
    );
  }
}

function buildDownloadUrl(platform: SupportedPlatform, version: string): string {
  // Map platform to release asset naming convention
  const assetMap: Record<SupportedPlatform, string> = {
    "darwin-arm64": "quack-darwin-arm64",
    "darwin-x64": "quack-darwin-x64",
    "linux-x64": "quack-linux-x64",
    "linux-arm64": "quack-linux-arm64",
    "windows-x64": "quack-windows-x64.exe",
    "windows-arm64": "quack-windows-arm64.exe",
  };

  const assetName = assetMap[platform];
  return `${GITHUB_RELEASE_BASE}/${version}/${assetName}`;
}

async function downloadFile(url: string, targetPath: string): Promise<void> {
  const response = await fetch(url, {
    headers: {
      "User-Agent": "quack-search",
    },
  });

  if (!response.ok) {
    if (response.status === 404) {
      throw new QuackBinaryError(
        `Binary not found at ${url}`,
        "DOWNLOAD_NOT_FOUND",
        `The release may not exist or the platform is not supported.\n` +
          `Check available releases at: https://github.com/adistrim/quack/releases`
      );
    }
    throw new QuackBinaryError(
      `Failed to download binary: HTTP ${response.status}`,
      "DOWNLOAD_FAILED",
      `URL: ${url}`
    );
  }

  const body = response.body;
  if (!body) {
    throw new QuackBinaryError(
      "Empty response body",
      "DOWNLOAD_EMPTY"
    );
  }

  // Remove existing file if present
  if (existsSync(targetPath)) {
    unlinkSync(targetPath);
  }

  // Stream to file
  const fileStream = createWriteStream(targetPath);

  // Convert web stream to node stream for pipeline
  const reader = body.getReader();

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      fileStream.write(value);
    }
    fileStream.end();

    await new Promise<void>((resolve, reject) => {
      fileStream.on("finish", resolve);
      fileStream.on("error", reject);
    });
  } catch (err) {
    fileStream.close();
    if (existsSync(targetPath)) {
      unlinkSync(targetPath);
    }
    throw new QuackBinaryError(
      "Failed to write binary file",
      "DOWNLOAD_WRITE_FAILED",
      err instanceof Error ? err.message : String(err)
    );
  }
}

function getDefaultBinDir(): string {
  // Default to bin directory relative to this package
  const thisDir = dirname(new URL(import.meta.url).pathname);
  return join(thisDir, "..", "..", "bin");
}

/**
 * Check if a binary download is needed.
 * Returns true if no binary is currently available.
 */
export function needsDownload(): boolean {
  return !checkBinaryStatus().found;
}
