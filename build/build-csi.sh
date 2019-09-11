#!/usr/bin/env bash
set -e

cd ${GOPATH}/src/github.com/tommenx/storage/cmd/csi


export GOARCH="amd64"
export GOOS="linux"


go build -o plugin.csi.alibabacloud.com

cd ${GOPATH}/src/github.com/tommenx/storage/build/csi
mv ${GOPATH}/src/github.com/tommenx/storage/cmd/csi/plugin.csi.alibabacloud.com ./
docker build -t=storage.io/csi-lvmplugin ./

rm -rf plugin.csi.alibabacloud.com