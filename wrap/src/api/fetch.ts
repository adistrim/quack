import { runCore } from "../core/runner";

export async function fetchContent(
  url: string,
  timeoutMs?: number
): Promise<string> {
  if (!url) {
    throw new Error("url is required");
  }

  return runCore(
    {
      action: "fetch",
      url,
    },
    timeoutMs
  );
}
