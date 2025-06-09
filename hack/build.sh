#!/bin/bash

set -euo pipefail

set -x
go test -cover ./...
