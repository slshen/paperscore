#!/bin/bash

yr=$1
shift 1

set -euo pipefail
go run main.go tournament --us pride --re-matrix data/tweaked_re.csv "$@" \
    data/$yr/${yr}*.{gm,yaml}
