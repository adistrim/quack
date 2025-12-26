# Makefile

BIN_DIR := ../wrap/bin
CORE_DIR := core

.PHONY: all darwin-arm64 darwin-amd64 linux-amd64 windows-amd64 clean

all: darwin-arm64 darwin-amd64 linux-amd64 windows-amd64

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -C $(CORE_DIR) -o $(BIN_DIR)/quack-darwin-arm64

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -C $(CORE_DIR) -o $(BIN_DIR)/quack-darwin-x64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -C $(CORE_DIR) -o $(BIN_DIR)/quack-linux-x64

windows-amd64:
	GOOS=windows GOARCH=amd64 go build -C $(CORE_DIR) -o $(BIN_DIR)/quack-windows-x64.exe

clean:
	rm -f $(BIN_DIR)/quack-*
