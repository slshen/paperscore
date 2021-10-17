#!/bin/bash

set -euo pipefail

go run main.go export --us pride data/2021*.yaml
