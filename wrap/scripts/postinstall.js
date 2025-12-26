import { platform, arch } from "node:process";
import { chmodSync, existsSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const BIN_DIR = join(__dirname, "..", "bin");

function suffix() {
  if (platform === "darwin" && arch === "arm64") return "darwin-arm64";
  if (platform === "darwin" && arch === "x64") return "darwin-x64";
  if (platform === "linux" && arch === "x64") return "linux-x64";
  if (platform === "win32" && arch === "x64") return "windows-x64.exe";
  throw new Error(`Unsupported platform: ${platform} ${arch}`);
}

const binPath = join(BIN_DIR, `quack-${suffix()}`);

if (!existsSync(binPath)) {
  console.error(`Missing binary: ${binPath}`);
  process.exit(1);
}

chmodSync(binPath, 0o755);
console.log(`quack binary ready: ${binPath}`);
