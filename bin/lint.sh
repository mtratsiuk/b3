#!/usr/bin/env bash

set -exu

cd "$(dirname "$0")"/..

go test ./...
not_formatted=$(go fmt ./...)

if [ ! -z "$not_formatted" ]; then
  echo "Following files are not formatted:"
  echo "$not_formatted"
  echo "Please run 'go fmt' and commit"
  exit 1
fi
