#!/bin/bash

set -euo pipefail

set -x
go test -cover ./...

linter=golangci-lint
if [ -x ./bin/golangci-lint ]; then
  linter=./bin/golangci-lint
fi
$linter run -E stylecheck -E gosec -E goimports -E misspell -E gocritic \
      -E whitespace -E goprintffuncname