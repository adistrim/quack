CORE_DIR := core

PKG_JS := packages/quack-search
PKG_DARWIN_ARM64 := packages/quack-search-darwin-arm64
PKG_DARWIN_X64 := packages/quack-search-darwin-x64
PKG_LINUX_X64 := packages/quack-search-linux-x64
PKG_WINDOWS_X64 := packages/quack-search-windows-x64

.PHONY: build build-js build-bins publish clean

# ---- build ----

build: build-js build-bins

build-js:
	cd $(PKG_JS) && bun run build

build-bins: darwin-arm64 darwin-x64 linux-x64 windows-x64

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -C $(CORE_DIR) -o ../$(PKG_DARWIN_ARM64)/bin/quack

darwin-x64:
	GOOS=darwin GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_DARWIN_X64)/bin/quack

linux-x64:
	GOOS=linux GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_LINUX_X64)/bin/quack

windows-x64:
	GOOS=windows GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_WINDOWS_X64)/bin/quack.exe

# ---- publish ----

publish: build
	cd $(PKG_DARWIN_ARM64) && npm publish
	cd $(PKG_DARWIN_X64) && npm publish
	cd $(PKG_LINUX_X64) && npm publish
	cd $(PKG_WINDOWS_X64) && npm publish
	cd $(PKG_JS) && npm publish

# ---- clean ----

clean:
	rm -f $(PKG_DARWIN_ARM64)/bin/quack
	rm -f $(PKG_DARWIN_X64)/bin/quack
	rm -f $(PKG_LINUX_X64)/bin/quack
	rm -f $(PKG_WINDOWS_X64)/bin/quack.exe
