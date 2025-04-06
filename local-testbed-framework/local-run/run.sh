#!/bin/bash

set -x
set -e

# compile runner
cd ./runner
go build -o ../local-run/runner ./main.go
cd ..

# create log directory
mkdir -p ./local-run/logs

# run
./local-run/runner exec -v \
  --log-file "./local-run/logs/$(date +%s).log" \
  --config "./local-run/runner-config.yml" \
  --artifacts-directory-path "./local-run/artifacts" \
  --test-cases-directory-path "./testcases" \
  --clients-directory-path "./clients" \
  --clients "safari"
