#!/bin/bash

rm -rf .git

git init .

git add -A

git commit -am "reduce"

git remote add origin git@github.com:snail007/gmct.git

git push --mirror --force