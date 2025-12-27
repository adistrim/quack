# Quack Search

DuckDuckGo web search library for JS runtimes, written in Go

### Installation

```bash
bun add quack-search
# or
npm install quack-search
# or
yarn add quack-search
```

### Usage

```ts
import { search } from 'quack-search';
console.log(await search('golang', { maxResults: 5 }));

const page = await fetchContent("https://adistrim.in/now");

if (!page.success) {
  console.log("Blocked:", page.reason);
} else {
  console.log(page.text);
}
```

#### _Open to Contributions_

## License
MIT License. See [LICENSE](LICENSE) file for details.

