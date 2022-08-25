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

# linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-amd64.tar.gz" gmct && rm gmct

# linux arm 5
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-arm-v5.tar.gz" gmct && rm gmct

# linux arm 6
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-arm-v6.tar.gz" gmct && rm gmct

# linux arm 7
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-arm-v7.tar.gz" gmct && rm gmct

# linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && upx -9 gmct && tar zcfv "${RELEASE_DIR}/gmct-linux-arm64.tar.gz" gmct && rm gmct

# darwin amd64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o gmct ./cmd/gmct/ && tar zcfv "${RELEASE_DIR}/gmct-mac-amd64.tar.gz" gmct && rm gmct

# windows amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o gmct.exe ./cmd/gmct/ && upx -9 gmct.exe && tar zcfv "${RELEASE_DIR}/gmct-windows-amd64.tar.gz" gmct.exe && rm gmct.exe

