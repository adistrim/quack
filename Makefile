CORE_DIR := core

PKG_JS := packages/quack-search
PKG_DARWIN_ARM64 := packages/quack-search-darwin-arm64
PKG_DARWIN_X64 := packages/quack-search-darwin-x64
PKG_LINUX_X64 := packages/quack-search-linux-x64
PKG_LINUX_ARM64 := packages/quack-search-linux-arm64
PKG_WINDOWS_X64 := packages/quack-search-windows-x64
PKG_WINDOWS_ARM64 := packages/quack-search-windows-arm64

ROOT_PKG := .
PKGS := \
	$(ROOT_PKG) \
	$(PKG_JS) \
	$(PKG_DARWIN_ARM64) \
	$(PKG_DARWIN_X64) \
	$(PKG_LINUX_X64) \
	$(PKG_LINUX_ARM64) \
	$(PKG_WINDOWS_X64) \
	$(PKG_WINDOWS_ARM64)

.PHONY: build build-js build-bins publish clean version check-version

# ---- versioning ----

check-version:
ifndef VERSION
	$(error VERSION is required. Usage: make version VERSION=x.y.z)
endif

version: check-version
	@echo "Updating all packages to version $(VERSION)"
	@for pkg in $(PKGS); do \
		echo "→ $$pkg"; \
		( \
			cd $$pkg && \
			jq '.version = "$(VERSION)"' package.json > package.json.tmp && \
			mv package.json.tmp package.json \
		); \
	done

	@echo "Syncing optionalDependencies in quack-search"
	@( \
		cd $(PKG_JS) && \
		jq '.optionalDependencies |= with_entries(.value = "$(VERSION)")' \
			package.json > package.json.tmp && \
		mv package.json.tmp package.json \
	)

# ---- release ----

release: build copy-assets checksums sign-checksums gh-release

copy-assets: build-bins
	@mkdir -p .release-assets
	cp $(PKG_DARWIN_ARM64)/bin/quack .release-assets/quack-darwin-arm64
	cp $(PKG_DARWIN_X64)/bin/quack .release-assets/quack-darwin-x64
	cp $(PKG_LINUX_X64)/bin/quack .release-assets/quack-linux-x64
	cp $(PKG_LINUX_ARM64)/bin/quack .release-assets/quack-linux-arm64
	cp $(PKG_WINDOWS_X64)/bin/quack.exe .release-assets/quack-windows-x64.exe
	cp $(PKG_WINDOWS_ARM64)/bin/quack.exe .release-assets/quack-windows-arm64.exe

checksums:
	cd .release-assets && \
	shasum -a 256 * > checksums.txt

sign-checksums:
	cd .release-assets && \
	gpg --armor --detach-sign checksums.txt

gh-release: check-version
	gh release create v$(VERSION) \
		.release-assets/* \
		--title "v$(VERSION)"

# ---- build ----

build: build-js build-bins

build-js:
	cd $(PKG_JS) && bun run build

build-bins: darwin-arm64 darwin-x64 linux-x64 linux-arm64 windows-x64 windows-arm64

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -C $(CORE_DIR) -o ../$(PKG_DARWIN_ARM64)/bin/quack

darwin-x64:
	GOOS=darwin GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_DARWIN_X64)/bin/quack

linux-x64:
	GOOS=linux GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_LINUX_X64)/bin/quack

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -C $(CORE_DIR) -o ../$(PKG_LINUX_ARM64)/bin/quack

windows-x64:
	GOOS=windows GOARCH=amd64 go build -C $(CORE_DIR) -o ../$(PKG_WINDOWS_X64)/bin/quack.exe

windows-arm64:
	GOOS=windows GOARCH=arm64 go build -C $(CORE_DIR) -o ../$(PKG_WINDOWS_ARM64)/bin/quack.exe

# ---- publish ----

publish:
	cd $(PKG_DARWIN_ARM64) && npm publish
	cd $(PKG_DARWIN_X64) && npm publish
	cd $(PKG_LINUX_X64) && npm publish
	cd $(PKG_LINUX_ARM64) && npm publish
	cd $(PKG_WINDOWS_X64) && npm publish
	cd $(PKG_WINDOWS_ARM64) && npm publish
	cd $(PKG_JS) && npm publish

# ---- clean ----

clean:
	rm -f $(PKG_DARWIN_ARM64)/bin/quack
	rm -f $(PKG_DARWIN_X64)/bin/quack
	rm -f $(PKG_LINUX_X64)/bin/quack
	rm -f $(PKG_LINUX_ARM64)/bin/quack
	rm -f $(PKG_WINDOWS_X64)/bin/quack.exe
	rm -f $(PKG_WINDOWS_ARM64)/bin/quack.exe
