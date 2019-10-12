#!/bin/bash -e

cd $GOPATH/src/github.com/tommenx/storage/cmd/csi/
go build -o csi-plugin main.go
cd $GOPATH/src/github.com/tommenx/storage/cmd/catchup
go build -o catch-up main.go
cd $GOPATH/src/github.com/tommenx/storage/cmd/coordinator
go build -o coordinator main.go
cd $GOPATH/src/github.com/tommenx/storage/bin
mv ../cmd/csi/csi-plugin ./
mv ../cmd/catchup/catch-up ./
mv ../cmd/coordinator/coordinator ./






