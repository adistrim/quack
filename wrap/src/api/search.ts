import { runCore } from "../core/runner";
import type { SearchOptions } from "../types/core";

export async function search(
  query: string,
  options: SearchOptions = {}
): Promise<string> {
  if (!query) {
    throw new Error("query is required");
  }

  return runCore(
    {
      action: "search",
      query,
      maxResults: options.maxResults ?? 10,
    },
    options.timeoutMs
  );
}
