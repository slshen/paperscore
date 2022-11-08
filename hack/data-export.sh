#!/bin/bash

set -euo pipefail

set -x
rm -rf export
go run main.go data-export --re-matrix data/d3_softball_re_2022.csv \
    --us pride --games data -d export

(cd export
ls
aws s3 cp --recursive . s3://slshen-public-us-west-2/pride-jf)
