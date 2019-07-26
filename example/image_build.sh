#!/bin/bash

version=v0.0
repo="ccr.ccs.tencentyun.com/ivanscai/test"

case "$1" in
build)
        docker build -t $repo:$version .
        ;;
push)
        export http_proxy=http://dev-proxy.oa.com:8080;export https_proxy=$http_proxy;export HTTP_PROXY=$http_proxy;export HTTPS_PROXY=$http_proxy
        docker push $repo:$version
        ;;
*)
        echo "usage:$0 [build|push]\n"
        exit 1
esac
exit 0