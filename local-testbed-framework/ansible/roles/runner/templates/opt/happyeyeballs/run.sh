#!/bin/bash

set -x
set -e

# compile runner
cd /opt/happyeyeballs/runner
/usr/local/go/bin/go build -o runner ./main.go

# create log directory
mkdir -p /opt/happyeyeballs/logs

# run
/opt/happyeyeballs/runner/runner exec -v \
  --log-file "/opt/happyeyeballs/logs/$(date +%s).log" \
  --config "/opt/happyeyeballs/runner-config.yml" \
  --artifacts-directory-path "/opt/happyeyeballs/artifacts" \
  --test-cases-directory-path "/opt/happyeyeballs/testcases" \
  --clients-directory-path "/opt/happyeyeballs/clients"

# upload logs and artifacts
cd /opt/happyeyeballs
pos_upload -r -f logs
pos_upload -r -f artifacts
