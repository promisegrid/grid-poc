# Makefile for building the GopherJS demo

GOPHERJS = $(shell which gopherjs)

all: build

build:
	$(GOPHERJS) build hello.go -o hello.js

clean:
	rm -f hello.js

.PHONY: all build clean
