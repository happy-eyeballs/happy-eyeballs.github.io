#!/bin/bash

tshark -r $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.pcap -T json > $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.json

python3 $LC_BASE_DIR/evaluate.py $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.json
