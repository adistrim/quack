import { spawn } from "node:child_process";
import type { CoreRequest, CoreResponse } from "../types/core";
import { QuackBinaryError } from "./binary";
import { ensureBinaryReady } from "./download";

export class QuackRuntimeError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly details?: string
  ) {
    super(message);
    this.name = "QuackRuntimeError";
  }
}

/**
 * Runs the core binary with auto-bootstrap.
 * Binary is automatically downloaded if not present.
 */
export async function runCore(
  payload: CoreRequest,
  timeoutMs = 30_000
): Promise<CoreResponse> {
  // Auto-bootstrap: ensures binary exists, downloads if needed
  let binaryPath: string;
  try {
    binaryPath = await ensureBinaryReady();
  } catch (err) {
    if (err instanceof QuackBinaryError) {
      throw err;
    }
    throw new QuackRuntimeError(
      "Failed to prepare quack binary",
      "BOOTSTRAP_FAILED",
      err instanceof Error ? err.message : String(err)
    );
  }

  // Execute the binary
  return executeCore(binaryPath, payload, timeoutMs);
}

function executeCore(
  binaryPath: string,
  payload: CoreRequest,
  timeoutMs: number
): Promise<CoreResponse> {
  return new Promise((resolve, reject) => {
    const proc = spawn(binaryPath, {
      stdio: ["pipe", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    const timeout = setTimeout(() => {
      proc.kill();
      reject(
        new QuackRuntimeError(
          `Core process timed out after ${timeoutMs}ms`,
          "TIMEOUT",
          `Binary: ${binaryPath}`
        )
      );
    }, timeoutMs);

    proc.stdout.on("data", (d: Buffer) => {
      stdout += d.toString();
    });

    proc.stderr.on("data", (d: Buffer) => {
      stderr += d.toString();
    });

    proc.on("error", (err) => {
      clearTimeout(timeout);

      // Handle common spawn errors with actionable messages
      if ((err as NodeJS.ErrnoException).code === "ENOENT") {
        reject(
          new QuackRuntimeError(
            `Binary not found or not executable: ${binaryPath}`,
            "ENOENT",
            `The binary path was resolved but the file doesn't exist or isn't accessible.\n` +
              `Check that the binary exists and has execute permissions.`
          )
        );
        return;
      }

      if ((err as NodeJS.ErrnoException).code === "EACCES") {
        reject(
          new QuackRuntimeError(
            `Permission denied executing binary: ${binaryPath}`,
            "EACCES",
            `Run: chmod +x "${binaryPath}"`
          )
        );
        return;
      }

      reject(
        new QuackRuntimeError(
          `Failed to spawn core process: ${err.message}`,
          "SPAWN_ERROR",
          `Binary: ${binaryPath}`
        )
      );
    });

    proc.on("close", (code) => {
      clearTimeout(timeout);

      if (code !== 0) {
        reject(
          new QuackRuntimeError(
            stderr || `Core process exited with code ${code}`,
            "PROCESS_FAILED",
            `Exit code: ${code}, Binary: ${binaryPath}`
          )
        );
        return;
      }

      let parsed: CoreResponse;
      try {
        parsed = JSON.parse(stdout);
      } catch {
        reject(
          new QuackRuntimeError(
            "Invalid JSON returned from core",
            "INVALID_RESPONSE",
            `Raw output: ${stdout.slice(0, 200)}${stdout.length > 200 ? "..." : ""}`
          )
        );
        return;
      }

      if (parsed.error) {
        reject(new QuackRuntimeError(parsed.error, "CORE_ERROR"));
        return;
      }

      if (!parsed.results && !parsed.fetch) {
        reject(
          new QuackRuntimeError("Empty response from core", "EMPTY_RESPONSE")
        );
        return;
      }

      resolve(parsed);
    });

    proc.stdin.write(JSON.stringify(payload));
    proc.stdin.end();
  });
}
