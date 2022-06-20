#!/bin/bash

set -euo pipefail

set -x
go run main.go export --re-matrix data/tweaked_re.csv \
    --us pride --spreadsheet-id 1th08uUbelfEnnxMfBhcvELdk4wd48rYxU52q-E5atbg \
    data/2022/2022*.{yaml,gm}