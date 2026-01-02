# Quack Search

DuckDuckGo web search library for JS runtimes, written in Go

## Installation

```bash
bun add quack-search
# or
npm install quack-search
# or
yarn add quack-search
```

## Usage

```ts
import { search, fetchContent } from 'quack-search';
console.log(await search('golang', { maxResults: 5 }));

const page = await fetchContent("https://adistrim.in/now");

if (!page.success) {
  console.log("Blocked:", page.reason);
} else {
  console.log(page.text);
}
```

## Supported Platforms

- macOS (x64, ARM64)
- Linux (x64, ARM64)
- Windows (x64, ARM64)

## Advanced: CI / Docker / Offline Environments

For constrained environments where automatic download isn't possible:

### Option 1: Environment Variable

Pre-install the binary and point to it:
```bash
export QUACK_BINARY_PATH=/usr/local/bin/quack
```

### Option 2: Explicit Download

Download at build time (e.g., in a Dockerfile):
```ts
import { downloadBinary } from 'quack-search';

await downloadBinary({ 
  targetDir: '/app/bin',
  version: 'v0.1.0', // or 'latest'
  verbose: true 
});
```

### Option 3: Platform Package

Install the platform-specific package directly:
```bash
npm install quack-search-linux-x64
```

### Diagnostics

```ts
import { checkBinaryStatus } from 'quack-search';

const status = checkBinaryStatus();
console.log(status);
// { found: true, path: '/path/to/quack', source: 'node_modules', platform: 'darwin', arch: 'arm64' }
```

## Contributing

Open to contributions! See the [repository](https://github.com/adistrim/quack).

## License
MIT License. See [LICENSE](LICENSE) file for details.

