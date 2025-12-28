/**

SIMPLE TEST SCRIPT

node test.js
or
bun test.js

*/


import {
  search,
  fetchContent,
  checkBinaryStatus,
  getPlatformTarget
} from "./packages/quack-search/dist/index.js";

console.log("=== Environment ===");
console.log("Platform target:", getPlatformTarget());

// Optional: diagnostic only (should NOT gate execution)
console.log("\n=== Binary Status (pre-run) ===");
const statusBefore = checkBinaryStatus();
console.log(JSON.stringify(statusBefore, null, 2));

console.log("\n=== Search Test (auto-bootstrap) ===");

try {
  const results = await search("adistrim", { maxResults: 3 });

  for (const r of results) {
    console.log(r.title, r.url);
  }
} catch (err) {
  console.error("Search failed:", err);
  process.exit(1);
}

// Binary should now exist (either npm, cache, or override)
console.log("\n=== Binary Status (post-run) ===");
const statusAfter = checkBinaryStatus();
console.log(JSON.stringify(statusAfter, null, 2));

console.log("\n=== Fetch Test ===");

try {
  const page = await fetchContent("https://adistrim.in/now");

  if (!page.success) {
    console.log("Blocked:", page.reason);
  } else {
    console.log(page.text.slice(0, 500));
  }
} catch (err) {
  console.error("Fetch failed:", err);
  process.exit(1);
}

console.log("\nAll tests completed successfully");
