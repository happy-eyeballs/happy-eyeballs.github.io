#!/bin/bash

export PATH=$PATH:/usr/local/bin

NONCE=$(cat /dev/urandom | base32 | head -c 6)

$LC_CLIENT_DIR/request.sh "id-$NONCE.delay_aaaa-$LC_DNS_DELAY.dns-delay-wg.$LC_BASE_DOMAIN" AAAA
