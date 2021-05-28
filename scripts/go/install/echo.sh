#!/bin/bash
set -e
# This is a standard gmct script example, you should only to
# implements function install, uninstall, installed.

install() {
  # if success, exit 0, others exit 1.
  # do something to install.
  echo "It works!"
  exit 0
}

uninstall() {
  # if success, exit 0, others exit 1.
  # do something to uninstall
  exit 0
}

installed() {
  # if installed, exit 0, others exit 1
  exit 0
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
