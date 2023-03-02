#!/bin/bash
set -e
# This is a standard gmct script example, you should only to
# implements function install, uninstall, installed.

install() {
  # if success, exit 0, others exit 1.
  # do something to install.
cat >/usr/local/bin/chgo<<EOF
#!/usr/bin/env bash
GO_VER_DIR="/usr/local/"
GO_ROOT="/usr/local/go"
VER="\$1"

if [ -z "\$VER" ]; then
  echo -e "chgo <go-main-version>\nExample: chgo 1.18"
  exit
fi
if [ -e "\$GO_ROOT" ];then
  if [ ! -L "\$GO_ROOT" ];then
    echo "GO_ROOT: \$GO_ROOT is not a symbolic link."
    exit
  fi
fi
newVer=\$(ls \$GO_VER_DIR|grep "go\$VER"|tail -n 1)
if [ -z "\$newVer" ]; then
  echo -e "none go found in: \$GO_VER_DIR"
  exit
fi

sudo rm "\$GO_ROOT"
sudo ln -s "\$GO_VER_DIR\$newVer" \$GO_ROOT
echo "switch to \$(go version)"
EOF
chmod +x /usr/local/bin/chgo
  echo "install success at path: /usr/local/bin/chgo"
  exit 0
}

uninstall() {
  # if success, exit 0, others exit 1.
  # do something to uninstall
  rm -f /usr/local/bin/chgo
  echo "uninstall success"
  exit 0
}

installed() {
  # if installed, exit 1, others exit 0
  if [ -f "/usr/local/bin/chgo" ];then
    exit 1
  else
    exit 0
  fi
}

case $ACTION in
install)
  install
  ;;
uninstall)
  uninstall
  ;;
installed)
  installed
  ;;
esac