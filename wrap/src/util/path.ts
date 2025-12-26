import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

export function resolveDirname(metaUrl: string): string {
  return typeof (import.meta as any).dir === "string"
    ? (import.meta as any).dir // Bun
    : dirname(fileURLToPath(metaUrl)); // Node
}
