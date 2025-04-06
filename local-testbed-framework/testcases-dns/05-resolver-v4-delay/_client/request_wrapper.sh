#!/bin/bash

export PATH=$PATH:/usr/local/bin

NONCE=$(cat /dev/urandom | base32 | head -c 16)

$LC_CLIENT_DIR/request.sh "id-$NONCE.dns-delay.$LC_BASE_DOMAIN"
