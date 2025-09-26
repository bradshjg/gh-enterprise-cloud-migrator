#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

function dev {
    cd "$SCRIPT_DIR/src/ghec-migrator"
    templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
}

"$@"
