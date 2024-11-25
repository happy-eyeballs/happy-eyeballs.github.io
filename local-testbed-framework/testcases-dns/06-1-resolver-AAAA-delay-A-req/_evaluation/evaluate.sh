#!/bin/bash

# tshark -r $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.pcap -T json > $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.json

# DNS_A_ZONE=$(echo $LC_DNS_A | cut -d ' ' -f 2)
# DNS_AAAA_ZONE=$(echo $LC_DNS_AAAA | cut -d ' ' -f 2)

# python3 $LC_BASE_DIR/evaluate.py $LC_EVALUATION_ARTIFACTS_DIR/tcpdump/pcaps/capture.json ${DNS_A_ZONE} ${DNS_AAAA_ZONE}
