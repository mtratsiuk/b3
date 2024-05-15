#!/usr/bin/env bash

set -exu

cd "$(dirname "$0")"/..

go run ./cmd/b3/main.go --root=./example -v $@
