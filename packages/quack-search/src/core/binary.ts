import { platform, arch } from "node:process";
import { dirname, join } from "node:path";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);

function resolveBinaryPackage(): string {
  if (platform === "darwin" && arch === "arm64") {
    return "quack-search-darwin-arm64";
  }

  if (platform === "darwin" && arch === "x64") {
    return "quack-search-darwin-x64";
  }

  if (platform === "linux" && arch === "x64") {
    return "quack-search-linux-x64";
  }

  if (platform === "win32" && arch === "x64") {
    return "quack-search-windows-x64";
  }

  throw new Error(`unsupported platform: ${platform} ${arch}`);
}

export function resolveBinaryPath(): string {
  const pkgName = resolveBinaryPackage();
  const pkgEntry = require.resolve(pkgName);
  
  return pkgEntry;
}
