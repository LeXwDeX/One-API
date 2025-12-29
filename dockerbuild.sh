#!/bin/bash

# 配置参数
DOCKER_USER="sunsan05"
IMAGE_NAME="one-api"
TAG="latest"
FULL_IMAGE_NAME="${DOCKER_USER}/${IMAGE_NAME}:${TAG}"

echo "==== 0. 更新项目  ===="
git fetch
git reset --hard origin/main
git pull origin main
if [ $? -ne 0 ]; then
  echo "更新项目失败，请检查网络和权限。"
  exit 1
fi
echo "==== 0. 更新项目成功 ===="

echo "==== 1. 构建镜像: ${FULL_IMAGE_NAME} ===="
docker build -t ${FULL_IMAGE_NAME} .
if [ $? -ne 0 ]; then
  echo "镜像构建失败，请检查 Dockerfile 和构建上下文。"
  exit 1
fi

echo "==== 2. 推送镜像到 Docker Hub ===="
docker push ${FULL_IMAGE_NAME}
if [ $? -ne 0 ]; then
  echo "镜像推送失败，请检查网络和权限。"
  exit 2
fi

echo "==== 3. 镜像推送成功！===="
echo "你可以在 https://hub.docker.com/r/${DOCKER_USER}/${IMAGE_NAME}/tags 查看你的镜像。"
