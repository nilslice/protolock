#!/bin/bash

set +x
mkdir -p builds

rm builds/*

NOW=`date +%F.%s`
BUILD_SRC=cmd/protolock/main.go
NAME=protolock
BUILD_CMD="go build -o $NAME $BUILD_SRC"

LINUX_TAR="protolock-$NOW-linux.tgz"
WIN_TAR="protolock-$NOW-windows.tgz"
MAC_TAR="protolock-$NOW-macos.tgz"

# cross compile for linux, windows, macOS
GOOS=linux $BUILD_CMD
tar czf $LINUX_TAR protolock README.md LICENSE
rm protolock
mv $LINUX_TAR builds

GOOS=windows $BUILD_CMD 
mv protolock protolock.exe
tar czf $WIN_TAR  protolock.exe README.md LICENSE
rm protolock.exe
mv $WIN_TAR builds

GOOS=darwin $BUILD_CMD
tar czf $MAC_TAR protolock README.md LICENSE
rm protolock
mv $MAC_TAR builds
