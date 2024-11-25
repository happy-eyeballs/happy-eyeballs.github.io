#!/bin/bash

export PATH=$PATH:/usr/local/bin

NONCE=$(cat /dev/urandom | base32 | head -c 16)

$LC_CLIENT_DIR/request.sh "http://delay_aaaa-$LC_AAAA_DELAY.nonce-$NONCE.server.$LC_BASE_DOMAIN"
