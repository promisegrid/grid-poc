# Makefile for building the WebAssembly demo

GOOS=js
GOARCH=wasm
GOCMD=go
OUTPUT=hello.wasm

all: build

build:
	$(GOCMD) build -o $(OUTPUT) hello.go

clean:
	rm -f $(OUTPUT)

.PHONY: all build clean
