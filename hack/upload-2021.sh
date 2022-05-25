#!/bin/bash

set -euo pipefail

set -x
go run main.go export --re-matrix data/tweaked_re.csv \
    --us pride data/2021/2021*.yaml
go run main.go export --re-matrix data/tweaked_re.csv \
    --us pride --league PGF --spreadsheet-id 1p5tcdAD461Jq1TZC68KwhWSr_5q3AWMyVTJEsaQrLP4 \
    data/2021/2021*.yaml