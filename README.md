# Graphviz-Server

Graphviz-Server 是一个 Web 服务，封装了对 Graphviz 的接口调用，实现通过 Web API 的方式生成 Graphviz 图形。

Docker 启动

```bash
docker run -d --restart=always -p 19921:19921 -v /data/graphviz-data/:/root/graphviz-data mylxsw/graphviz-server:latest
```

支持生成文件类型：

    "svg", "svgz", "webp", "png", "bmp", "jpg", "jpeg", "pdf", "gif"

请求示例

```bash
curl --location --request POST 'http://127.0.0.1:19921/api/graphviz/?type=svg' \
--data-binary '@/Users/mylxsw/codes/examples/graphviz/nginx-flow.dot'
```

