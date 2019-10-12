#!/bin/bash -e

cd $GOPATH/src/github.com/tommenx/storage/cmd/csi/
go build -o csi-plugin main.go
cd $GOPATH/src/github.com/tommenx/storage/cmd/catchup
go build -o catch-up main.go

cd $GOPATH/src/github.com/tommenx/storage/bin
cp ../cmd/csi/csi-plugin ./
cp ../cmd/catchup/catch-up ./



