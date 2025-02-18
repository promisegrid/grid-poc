# Makefile for building the WASM Component Model Hello Demo

# Set environment variables for WebAssembly builds.
WASM_GOOS=js
WASM_GOARCH=wasm

# Default target: build and componentize.
all: component

# Build the standard WASM binary from hello.go.
hello.wasm: hello.go
	GOOS=$(WASM_GOOS) GOARCH=$(WASM_GOARCH) go build -o hello.wasm hello.go

# Componentize the WASM binary using wasm-tools.
# Note: Ensure that 'wasm-tools' is installed and available in your PATH.
component: hello.wasm hello.yaml hello.wit
	wasm-tools component new hello.wasm -c hello.yaml -o main.wasm

# Run a local HTTP server to serve the files (requires Python3).
run: all
	python3 -m http.server

# Clean build artifacts.
clean:
	rm -f hello.wasm main.wasm

.PHONY: all component run clean
