#!/bin/bash
F="gmct-$1.tar.gz"
set -e
if [ -e /tmp/gmct ]; then
    rm -rf /tmp/gmct
fi
mkdir /tmp/gmct
cd /tmp/gmct
echo -e "\n>>> downloading ... $F\n"

LAST_VERSION=$(curl --silent "https://mirrors.host900.com/https://api.github.com/repos/snail007/gmct/releases/latest" | grep -Po '"tag_name":"\K.*?(?=")')
wget  -t 1 "https://mirrors.host900.com/https://github.com/snail007/gmct/releases/download/${LAST_VERSION}/$F"

echo -e ">>> installing ... \n"

tar zxvf $F >/dev/null 2>&1
chmod +x gmct

p="/usr/bin"
if [ "$(uname)" == "Darwin" ] ;then
    p="/usr/local/bin"
    if [ ! -d "$p" ];then
      mkdir -p "$p"
    fi
fi

mv gmct "$p"

rm $F
gmct --version
echo  -e "\n>>> install success, thanks for using snail007/gmct\n"
echo  -e ">>> execute binary path: /usr/bin/gmct\n"

