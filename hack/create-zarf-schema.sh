#!/usr/bin/env sh

# Create the json schema for the jackal.yaml
go run main.go internal gen-config-schema > jackal.schema.json

# Adds pattern properties to all definitions to allow for yaml extensions
jq '.definitions |= map_values(. + {"patternProperties": {"^x-": {}}})' jackal.schema.json > temp_jackal.schema.json
mv temp_jackal.schema.json jackal.schema.json

# Create docs from the jackal.yaml JSON schema
docker run -v $(pwd):/app -w /app --rm python:3.8-alpine /bin/sh -c "pip install json-schema-for-humans && generate-schema-doc --config-file hack/.templates/jsfh-config.json jackal.schema.json docs/3-create-a-jackal-package/4-jackal-schema.md"
