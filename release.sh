#!/bin/bash
if [ -z "$1" ];then
echo "$0 0.0.1"
exit
fi

VERSION_RAW=$1
VERSION="v$VERSION_RAW"
RELEASE_DIR="release-$VERSION"

cat >cmd/gmct/version.go <<EOF
package main
var version = "$VERSION_RAW"
EOF

rm -rf $RELEASE_DIR
mkdir $RELEASE_DIR

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-amd64.tar.gz" gmct && rm gmct

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && tar zcfv "${RELEASE_DIR}/gmct-mac-amd64.tar.gz" gmct && rm gmct

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o gmct.exe ./cmd/gmct/ && upx -9 gmct.exe && tar zcfv "${RELEASE_DIR}/gmct-windows-amd64.tar.gz" gmct.exe && rm gmct.exe

