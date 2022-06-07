#!/bin/bash
# $1:harbor用户名   $2:harbor密码

user=$1
pwd=$2
docker logout harbor.cmft.com
docker login -u${user} -p${pwd} harbor.cmft.com