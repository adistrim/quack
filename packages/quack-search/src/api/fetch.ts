import { runCore } from "../core/runner";
import type { FetchResult } from "../types/core";

export async function fetchContent(
  url: string,
  timeoutMs?: number
): Promise<FetchResult> {
  const res = await runCore({
    action: "fetch",
    url,
  }, timeoutMs);

  return res.fetch ?? { success: false, reason: "unknown" };
}
