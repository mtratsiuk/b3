#!/usr/bin/env bash

set -exu

cd "$(dirname "$0")"/..

last_hash=""

inotifywait -m -r ./example/posts ./example/assets ./example/b3.json ./pkg -e create -e delete -e move -e modify |
    while read directory action file; do
        changed="$directory$file"

        if [ -e "$changed" ]; then
            current_hash=$(md5sum "$changed")
        else
            current_hash=$changed
        fi

        if [ "$current_hash" != "$last_hash" ]; then
            echo "changed"
            ./bin/run_example.sh
        fi

        last_hash="$current_hash"
    done
