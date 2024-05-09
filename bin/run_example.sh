#!/usr/bin/env bash

set -exu

cd "$(dirname "$0")"/..

go run ./cmd/cli/cli.go --root=./example -v
