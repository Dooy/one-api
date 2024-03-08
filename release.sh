#!/bin/bash

set -e

docker build  -t ydlhero/myone:latest .
# 修改镜像标签为当前日期时间
time=$(date "+%Y%m%d%H%M%S")
# 获取当前git commit id 
commit_id=$(git rev-parse HEAD)
docker tag ydlhero/myone:latest ydlhero/myone:$time-$commit_id
# 推送镜像到docker hub
docker push ydlhero/myone:$time-$commit_id
docker push ydlhero/myone:latest