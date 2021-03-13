#!/bin/bash

rm -rf /tmp/rbox_build || true
mkdir -pv /tmp/rbox_build
mkdir -pv /tmp/rbox_build/DEBIAN
mkdir -pv /tmp/rbox_build/usr/bin

go version
go build -v -o /tmp/rbox_build/usr/bin/rbox rbox/main.go
cp ci/control       /tmp/rbox_build/DEBIAN/control

dpkg-deb --build /tmp/rbox_build ./
