#!/bin/bash

set -euo pipefail
go run main.go tournament --us pride --re-matrix data/tweaked_re.csv "$@" \
    data/2021*.yaml
