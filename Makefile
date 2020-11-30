Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := -s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)

run: build
	./build/debug/graphviz-server --debug

build:
	go build -race -ldflags "$(LDFLAGS)" -o build/debug/graphviz-server main.go

esc-build: build-dashboard
	esc -pkg assets -o assets/static.go -prefix=assets assets
	esc -pkg dashboard -o dashboard/dashboard.go -prefix=dashboard/dist dashboard/dist

build-dashboard:
	cd dashboard && yarn build

docker-build:
	docker build -t mylxsw/graphviz-server .

docker-push:
	sh ./docker-push.sh

.PHONY: run build docker-build esc-build
