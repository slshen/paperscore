#!/bin/bash

set -euo pipefail

go run main.go alt --re-matrix data/tweaked_re.csv "$@"