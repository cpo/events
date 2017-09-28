#!/bin/bash

DIR=`pwd`
export GOPATH=$DIR
export GOROOT=

echo Compiling ARM7...
go build -cgo ./src/github.com/cpo/events
mv events ./events.arm7


