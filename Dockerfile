FROM golang:1.14 AS server-build
RUN mkdir -p /golang/graphviz-server
WORKDIR /golang/graphviz-server
RUN go get -u github.com/mjibson/esc
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN esc -pkg api -o api/static.go -prefix=assets assets
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -X main.Version=latest -X main.GitCommit=24130b9704a9cd398932c3f0d2262b8568e02e65' -o graphviz-server main.go

FROM ubuntu:20.10
WORKDIR /root
RUN apt-get update && apt-get install -y graphviz --no-install-recommends && rm -r /var/lib/apt/lists/*
RUN mkdir -p /root/graphviz-data
COPY --from=server-build /golang/graphviz-server/graphviz-server .
EXPOSE 19921
CMD ["./graphviz-server", "--listen", ":19921", "--dot-bin", "/usr/bin/dot", "--tmpdir", "/root/graphviz-data"]