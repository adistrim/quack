import { platform, arch, env } from "node:process";
import { join, dirname, resolve } from "node:path";
import { existsSync, accessSync, constants } from "node:fs";
import { fileURLToPath } from "node:url";
import { homedir } from "node:os";

// ---------- Platform Detection ----------

export type SupportedPlatform =
  | "darwin-arm64"
  | "darwin-x64"
  | "linux-x64"
  | "linux-arm64"
  | "windows-x64"
  | "windows-arm64";

export function getPlatformTarget(): SupportedPlatform {
  const key = `${platform}-${arch}`;

  const supported: Record<string, SupportedPlatform> = {
    "darwin-arm64": "darwin-arm64",
    "darwin-x64": "darwin-x64",
    "linux-x64": "linux-x64",
    "linux-arm64": "linux-arm64",
    "win32-x64": "windows-x64",
    "win32-arm64": "windows-arm64",
  };

  const target = supported[key];
  if (!target) {
    throw new QuackBinaryError(
      `Unsupported platform: ${platform} ${arch}`,
      "UNSUPPORTED_PLATFORM",
      `quack-search supports: darwin-arm64, darwin-x64, linux-x64, linux-arm64, windows-x64, windows-arm64.\n` +
        `You can provide a custom binary via QUACK_BINARY_PATH environment variable.`
    );
  }

  return target;
}

export function getBinaryName(): string {
  return platform === "win32" ? "quack.exe" : "quack";
}

// ---------- Error Handling ----------

export class QuackBinaryError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly hint?: string
  ) {
    super(message);
    this.name = "QuackBinaryError";
  }

  override toString(): string {
    let msg = `${this.name}: ${this.message}`;
    if (this.hint) {
      msg += `\n\nHint: ${this.hint}`;
    }
    return msg;
  }
}

// ---------- Binary Resolution ----------

interface BinaryResolutionResult {
  path: string;
  source: "env" | "node_modules" | "local";
}

/**
 * Attempts to resolve the quack binary using multiple strategies:
 * 1. QUACK_BINARY_PATH environment variable (explicit override)
 * 2. Platform-specific npm package in node_modules
 * 3. Local binary in package directory (for development/bundled)
 */
export function resolveBinaryPath(): string {
  const result = tryResolveBinary();
  return result.path;
}

export function tryResolveBinary(): BinaryResolutionResult {
  // Explicit env var override
  const envPath = env.QUACK_BINARY_PATH;
  if (envPath) {
    if (!existsSync(envPath)) {
      throw new QuackBinaryError(
        `Binary not found at QUACK_BINARY_PATH: ${envPath}`,
        "ENV_BINARY_NOT_FOUND",
        `Ensure the file exists and is executable.`
      );
    }
    assertExecutable(envPath);
    return { path: envPath, source: "env" };
  }

  // Looking in node_modules for platform-specific package
  const nodeModulesPath = tryResolveFromNodeModules();
  if (nodeModulesPath) {
    return { path: nodeModulesPath, source: "node_modules" };
  }

  // Look for local binary
  const localPath = tryResolveLocalBinary();
  if (localPath) {
    return { path: localPath, source: "local" };
  }

  // all failure state
  const targetPlatform = getPlatformTarget();
  const packageName = `quack-search-${targetPlatform}`;

  throw new QuackBinaryError(
    `Could not find quack binary for ${platform} ${arch}`,
    "BINARY_NOT_FOUND",
    `To fix this, try one of:\n` +
      `  1. Install the platform package: npm install ${packageName}\n` +
      `  2. Set QUACK_BINARY_PATH to point to your quack binary\n` +
      `  3. Download the binary from: https://github.com/adistrim/quack/releases\n\n` +
      `If you're using Docker or a bundler, set QUACK_BINARY_PATH explicitly.`
  );
}

// resolve binary from platform-specific package in node_modules
function tryResolveFromNodeModules(): string | null {
  const targetPlatform = getPlatformTarget();
  const packageName = `quack-search-${targetPlatform}`;
  const binaryName = getBinaryName();
  const thisDir = getModuleDir();

  //.  node_modules locations to check
  const searchPaths = [
    // hoisted in monorepo (relative to this package)
    join(thisDir, "..", "..", "..", "node_modules", packageName, "bin", binaryName),
    // direct
    join(thisDir, "..", "..", "node_modules", packageName, "bin", binaryName),
    // peer location
    join(thisDir, "..", "..", packageName, "bin", binaryName),
    // pnpm-style
    join(thisDir, "..", "..", "..", ".pnpm", "node_modules", packageName, "bin", binaryName),
  ];

  // Also check process.cwd() based paths for runtime resolution
  const cwd = process.cwd();
  searchPaths.push(
    join(cwd, "node_modules", packageName, "bin", binaryName),
    join(cwd, "..", "node_modules", packageName, "bin", binaryName)
  );

  for (const searchPath of searchPaths) {
    const resolved = resolve(searchPath);
    if (existsSync(resolved)) {
      try {
        assertExecutable(resolved);
        return resolved;
      } catch {
        // next
      }
    }
  }

  return null;
}

function tryResolveLocalBinary(): string | null {
  const binaryName = getBinaryName();
  const thisDir = getModuleDir();

  // Check for local bin directory (useful for development or bundled distributions)
  const localPaths = [
    join(thisDir, "..", "..", "bin", binaryName),
    join(thisDir, "..", "bin", binaryName),
    join(thisDir, "bin", binaryName),
  ];

  // Also check user cache directory (auto-downloaded binaries)
  const cacheDir = getCacheDir();
  localPaths.push(join(cacheDir, binaryName));

  for (const localPath of localPaths) {
    const resolved = resolve(localPath);
    if (existsSync(resolved)) {
      try {
        assertExecutable(resolved);
        return resolved;
      } catch {
        // Not executable, try next
      }
    }
  }

  return null;
}

function getCacheDir(): string {
  const home = homedir();
  if (platform === "win32") {
    return join(env.LOCALAPPDATA || join(home, "AppData", "Local"), "quack-search");
  }
  return join(env.XDG_CACHE_HOME || join(home, ".cache"), "quack-search");
}

function assertExecutable(filePath: string): void {
  try {
    accessSync(filePath, constants.X_OK);
  } catch {
    throw new QuackBinaryError(
      `Binary exists but is not executable: ${filePath}`,
      "BINARY_NOT_EXECUTABLE",
      `Run: chmod +x "${filePath}"`
    );
  }
}

function getModuleDir(): string {
  // Support both Bun and Node.js
  if (typeof (import.meta as any).dir === "string") {
    return (import.meta as any).dir;
  }
  return dirname(fileURLToPath(import.meta.url));
}

// ---------- Binary Status Check ----------

export interface BinaryStatus {
  found: boolean;
  path?: string;
  source?: "env" | "node_modules" | "local";
  error?: string;
  platform: string;
  arch: string;
}

/**
 * Check if the binary is available without throwing.
 * Useful for diagnostics and graceful degradation.
 */
export function checkBinaryStatus(): BinaryStatus {
  const base = { platform, arch };

  try {
    const result = tryResolveBinary();
    return {
      ...base,
      found: true,
      path: result.path,
      source: result.source,
    };
  } catch (err) {
    return {
      ...base,
      found: false,
      error: err instanceof Error ? err.message : String(err),
    };
  }
}

/**
 * Ensures the binary is available. Throws QuackBinaryError if not.
 * Call this at startup or before critical operations to fail fast.
 * 
 * @example
 * ```ts
 * import { ensureBinary } from 'quack-search';
 * 
 * // Fail fast at app startup
 * ensureBinary();
 * ```
 */
export function ensureBinary(): void {
  const status = checkBinaryStatus();
  if (!status.found) {
    const targetPlatform = getPlatformTarget();
    const packageName = `quack-search-${targetPlatform}`;

    throw new QuackBinaryError(
      `Quack binary not available for ${status.platform} ${status.arch}`,
      "BINARY_NOT_AVAILABLE",
      `To fix this:\n` +
        `  1. Install platform package: npm install ${packageName}\n` +
        `  2. Or set QUACK_BINARY_PATH environment variable\n` +
        `  3. Or download explicitly: await downloadBinary()\n\n` +
        `For Docker/CI: Set QUACK_BINARY_PATH or call downloadBinary() at build time.`
    );
  }
}
