#!/bin/bash
# $1:路径  $2:镜像名

path=$1
image=$2

# 创建临时目录
mkdir -p /tmp/caas/workspace/${path}
cd /tmp/caas/workspace/${path}

# 下载镜像并推送镜像
newImage=$(echo $image | sed 's#harbor.uat.cmft.com#harbor.cmft.com#g')
docker pull $image
docker tag $image $newImage
docker push $newImage
docker rmi $image
docker rmi $newImage

# 删除临时目录
cd /tmp/caas/workspace/
rm -rf ${path}