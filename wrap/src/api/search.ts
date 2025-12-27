import { runCore } from "../core/runner";
import type { SearchOptions, SearchResult } from "../types/core";

export async function search(
  query: string,
  options: SearchOptions = {}
): Promise<SearchResult[]> {
  const res = await runCore({
    action: "search",
    query,
    maxResults: options.maxResults ?? 10,
  }, options.timeoutMs);

  return res.results ?? [];
}
