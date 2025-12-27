// ---------- Public options ----------

export type SearchOptions = {
  maxResults?: number;
  timeoutMs?: number;
};

// ---------- Requests sent to Go core ----------

export type CoreRequest =
  | {
      action: "search";
      query: string;
      maxResults?: number;
    }
  | {
      action: "fetch";
      url: string;
    };

// ---------- Search ----------

export type SearchResult = {
  title: string;
  url: string;
  snippet: string;
  rank: number;
};

// ---------- Fetch ----------

export type FetchFailureReason =
  | "blocked"
  | "timeout"
  | "http_error"
  | "unknown";

export type FetchResult = {
  success: boolean;
  text?: string;
  reason?: FetchFailureReason;
  truncated?: boolean;
};

// ---------- Core response ----------

export type CoreResponse = {
  results?: SearchResult[];
  fetch?: FetchResult;
  error?: string;
};
