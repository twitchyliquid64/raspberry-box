#!/bin/bash
set -e

SCRIPT_BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
BUILD_DIR=${1%/}

if [ -f "${BUILD_DIR}" ]; then
	echo "Cannot build into '${BUILD_DIR}': it is a file."
	exit 1
fi

if [ -d "${BUILD_DIR}" ]; then
  rm -rfv ${BUILD_DIR}/*
fi

mkdir -pv ${BUILD_DIR}

go build -o "${BUILD_DIR}/usr/bin/rbox" github.com/twitchyliquid64/raspberry-box/rbox

mkdir -pv "${BUILD_DIR}/DEBIAN"
cp -rv ${SCRIPT_BASE_DIR}/DEBIAN/* "${BUILD_DIR}/DEBIAN"

dpkg-deb --build "${BUILD_DIR}" ./
