#!/bin/bash

docker buildx build \
    --platform="linux/amd64,linux/arm64" \
    -t paravirtualtishu/controller:1.3 \
    --push \
    .
