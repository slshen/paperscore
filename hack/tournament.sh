#!/bin/bash

dir=$1
shift 1

set -euo pipefail
go run main.go tournament --us pride --re-matrix data/tweaked_re.csv "$@" \
    data/$dir/*.yaml
