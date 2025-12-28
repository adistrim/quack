// ---------- Primary API ----------
// These are all most users need. Binary is managed automatically.

export { search } from "./api/search";
export { fetchContent } from "./api/fetch";

export type { SearchOptions, SearchResult, FetchResult } from "./types/core";

// ---------- Advanced: Binary Management ----------
// For CI, Docker, offline environments, or debugging.
// Most users should never need these.

export {
  checkBinaryStatus,
  ensureBinary,
  resolveBinaryPath,
  QuackBinaryError,
  getPlatformTarget,
  type BinaryStatus,
  type SupportedPlatform,
} from "./core/binary";

export { QuackRuntimeError } from "./core/runner";

export {
  downloadBinary,
  needsDownload,
  type DownloadOptions,
  type DownloadResult,
} from "./core/download";
