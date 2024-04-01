#!/usr/bin/env sh

if [ -z "$(git status -s docs/ jackal.schema.json)" ]; then
    echo "Success!"
    exit 0
else
    git diff docs/ jackal.schema.json
    exit 1
fi
