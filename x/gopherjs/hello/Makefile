# Makefile for building the GopherJS demo

GOPHERJS = $(shell which gopherjs)
export GOPHERJS_GOROOT = /home/stevegt/.goenv/versions/1.19.13

all: build

build:
	$(GOPHERJS) build hello.go -o hello.js

clean:
	rm -f hello.js

.PHONY: all build clean
