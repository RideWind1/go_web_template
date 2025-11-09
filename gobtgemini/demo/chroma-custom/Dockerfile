# 使用官方的 chroma 镜像作为我们的基础
FROM chromadb/chroma:0.4.24

# 安装 curl 工具。
# 大部分精简镜像使用 Alpine Linux，其包管理器是 apk。
# 这条命令会更新包列表，然后安装 curl，并且不保留缓存以减小体积。
RUN apt-get update && apt-get install -y curl