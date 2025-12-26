import { spawn } from "node:child_process";
import type { CoreRequest, CoreResponse } from "../types/core";
import { resolveBinaryPath } from "./binary";

export function runCore(
  payload: CoreRequest,
  timeoutMs = 30_000,
  metaUrl = import.meta.url
): Promise<string> {
  return new Promise((resolve, reject) => {
    const binaryPath = resolveBinaryPath(metaUrl);

    const proc = spawn(binaryPath, {
      stdio: ["pipe", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    const timeout = setTimeout(() => {
      proc.kill();
      reject(new Error("core process timed out"));
    }, timeoutMs);

    proc.stdout.on("data", (d) => {
      stdout += d.toString();
    });

    proc.stderr.on("data", (d) => {
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

      if (!parsed.text) {
        reject(new Error("empty response from core"));
        return;
      }

      resolve(parsed.text);
    });

    proc.stdin.write(JSON.stringify(payload));
    proc.stdin.end();
  });
}
