#!/usr/bin/env bash

# $1: 应用类型
# $2：发布单元名
# $3: 运行环境
# $4: vpc

if [ ! -d /app/appsystems/${1}_${2}_${3}_${4}/apps ]; then
  echo "app dir dose not exist"
  exit 1
fi

#if [ ! -d /app/logs/rtlog/${1}_${2}_${3}_${4} ]; then
#  echo "app logs dir dose not exist"
#  exit 1
#fi

if [ ! -d /app/backup ]; then
  echo "app backup dir dose not exist"
  exit 1
fi