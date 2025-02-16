# Makefile for building the WebAssembly demo

# Set GOOS and GOARCH for WebAssembly
WASM_GOOS=js
WASM_GOARCH=wasm

# Update GOROOT if necessary; otherwise, it is picked up from the environment
GOROOT ?= $(shell go env GOROOT)

# Default target builds the wasm binary and copies the wasm_exec.js file.
all: build wasm_exec.js

# Build hello.wasm from hello.go
build: hello.wasm

hello.wasm: hello.go
	GOOS=$(WASM_GOOS) GOARCH=$(WASM_GOARCH) go build -o hello.wasm hello.go

# Copy wasm_exec.js from the Go distribution to the current directory.
wasm_exec.js: $(GOROOT)/misc/wasm/wasm_exec.js
	cp $< $@

run: all
	go run ../server.go .

clean:
	rm -f hello.wasm wasm_exec.js

.PHONY: all build clean
