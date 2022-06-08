#!/bin/bash

set -euo pipefail

go run main.go re --raw data/202*/202*.yaml > data/freq.txt