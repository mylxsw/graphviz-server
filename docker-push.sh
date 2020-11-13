#!/usr/bin/env bash

TAG=`cat VERSION`

docker build -t mylxsw/graphviz-server .

docker tag mylxsw/graphviz-server mylxsw/graphviz-server:$TAG
docker tag mylxsw/graphviz-server:$TAG mylxsw/graphviz-server:latest
docker push mylxsw/graphviz-server:$TAG
docker push mylxsw/graphviz-server:latest

