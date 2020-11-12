Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := -s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)

run: build
	./build/debug/graphviz-server

build:
	go build -race -ldflags "$(LDFLAGS)" -o build/debug/graphviz-server main.go

.PHONY: run build
