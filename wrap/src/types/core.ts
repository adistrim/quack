export type SearchOptions = {
  maxResults?: number;
  timeoutMs?: number;
};

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

export type CoreResponse = {
  text?: string;
  error?: string;
};
