#!/bin/bash
GOOS=linux GOARCH=amd64 go build
xz -9v goup
mv goup.xz bin/goup-linux-x64.xz
GOOS=darwin GOARCH=amd64 go build
xz -9v goup
mv goup.xz bin/goup-macos-x64.xz
