#!/bin/bash

set -euo pipefail

go run main.go export --us pride data/2021*.yaml
go run main.go export --us pride --league PGF --spreadsheet-id 1p5tcdAD461Jq1TZC68KwhWSr_5q3AWMyVTJEsaQrLP4 data/2021*.yaml