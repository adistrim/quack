import { spawn } from "node:child_process";
import type { CoreRequest, CoreResponse } from "../types/core";
import { resolveBinaryPath } from "./binary";

export function runCore(
  payload: CoreRequest,
  timeoutMs = 30_000
): Promise<CoreResponse> {
  return new Promise((resolve, reject) => {
    let binaryPath: string;

    try {
      binaryPath = resolveBinaryPath();
    } catch (err) {
      reject(err);
      return;
    }

    const proc = spawn(binaryPath, {
      stdio: ["pipe", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    const timeout = setTimeout(() => {
      proc.kill();
      reject(new Error("core process timed out"));
    }, timeoutMs);

    proc.stdout.on("data", (d: Buffer) => {
      stdout += d.toString();
    });

    proc.stderr.on("data", (d: Buffer) => {
      stderr += d.toString();
    });

    proc.on("error", (err) => {
      clearTimeout(timeout);
      reject(err);
    });

    proc.on("close", (code) => {
      clearTimeout(timeout);

      if (code !== 0) {
        reject(new Error(stderr || "core process failed"));
        return;
      }

      let parsed: CoreResponse;
      try {
        parsed = JSON.parse(stdout);
      } catch {
        reject(new Error("invalid JSON returned from core"));
        return;
      }

      if (parsed.error) {
        reject(new Error(parsed.error));
        return;
      }

      if (!parsed.results && !parsed.fetch) {
        reject(new Error("empty response from core"));
        return;
      }

      resolve(parsed);
    });

    proc.stdin.write(JSON.stringify(payload));
    proc.stdin.end();
  });
}
