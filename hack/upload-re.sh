#!/bin/bash

set -euo pipefail

go run main.go re --plain --csv data > data/observed_re.csv

for f in d1_softball_re_2019.csv \
    d1_softball_re_2022.csv \
    d2_softball_re_2022.csv \
    d3_softball_re_2022.csv \
    mlb_re_2010-2015.csv \
    observed_re.csv; do
    aws s3 cp data/$f s3://slshen-public-us-west-2/data/re/
done
