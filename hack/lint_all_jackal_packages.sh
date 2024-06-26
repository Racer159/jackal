#!/bin/bash

JACKAL_BIN=$1
LINT_SRC_TEST=$2
SCRIPT=$(realpath "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
cd "$SCRIPTPATH" || exit
cd ..
find "." -type f -name 'jackal.yaml' | while read -r yaml_file; do
  dir=$(dirname "$yaml_file")
  if [[ "$dir" == *src/test/* ]] && [ "$LINT_SRC_TEST" != true ]; then
      continue
  fi
  echo "Running 'jackal prepare lint' in directory: $dir"
  $JACKAL_BIN prepare lint "$dir"
  echo "---"
done
