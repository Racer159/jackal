#!/usr/bin/env sh

# Create the json schema for the zarf.yaml
go run main.go internal config-schema > zarf.schema.json

# Create the json schema for the API and use it to create the typescript definitions
go run main.go internal api-schema | npx quicktype -s schema -o src/ui/lib/api-types.ts

# Create docs from the zarf.yaml JSON schema
docker run -v $(pwd):/app -w /app --rm python:3.8-alpine /bin/sh -c "pip install json-schema-for-humans && generate-schema-doc --config-file docs/.templates/jsfh-config.json zarf.schema.json docs/3-create-a-zarf-package/4-zarf-schema.md"
