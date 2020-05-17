#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -o MessageServer ../

cid=`docker ps -a| grep MessageServer | awk '{print $1}'`
iid=`docker images| grep message-server | awk '{print $3}'`
if [ "X${cid}" != "X" ];then
    docker stop $cid && docker rm $cid
fi

if [ "X${iid}" != "X" ];then
    docker rmi $iid
fi

docker build -t message-server:latest .
rm MessageServer