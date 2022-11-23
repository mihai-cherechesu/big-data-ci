#!/bin/bash

docker buildx build \
    --platform="linux/amd64,linux/arm64,darwin/amd64,darwin/arm64" \
    -t paravirtualtishu/controller:1.2 \
    --push \
    .
