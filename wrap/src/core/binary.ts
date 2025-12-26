import { join } from "node:path";
import { resolveDirname } from "../util/path";

export function resolveBinaryPath(metaUrl: string): string {
  const platform = process.platform;
  const arch = process.arch;

  let suffix: string;

  if (platform === "darwin" && arch === "arm64") {
    suffix = "darwin-arm64";
  } else if (platform === "darwin" && arch === "x64") {
    suffix = "darwin-x64";
  } else if (platform === "linux" && arch === "x64") {
    suffix = "linux-x64";
  } else if (platform === "win32" && arch === "x64") {
    suffix = "windows-x64.exe";
  } else {
    throw new Error(`unsupported platform: ${platform} ${arch}`);
  }

  const dir = resolveDirname(metaUrl);
  return join(dir, "..", "bin", `quack-${suffix}`);
}
