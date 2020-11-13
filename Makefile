Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := -s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)

run: build
	./build/debug/graphviz-server --debug

build: esc-build
	go build -race -ldflags "$(LDFLAGS)" -o build/debug/graphviz-server main.go

esc-build:
	esc -pkg api -o api/static.go -prefix=assets assets

docker-build:
	docker build -t mylxsw/graphviz-server .

docker-push:
	sh ./docker-push.sh

.PHONY: run build docker-build esc-build
