#!/bin/bash

DOCKER_COMPOSE_CMD="docker compose -f $LC_BASE_DIR/docker-compose.yml"

# Substitute DNS zonefile
export BASE_DOMAIN=$LC_BASE_DOMAIN
export DNS_BIND_ADDRESS=$LC_DNS_BIND_ADDRESS
export DNS_SERIAL=$(date +%s)
export DNS_A=$(echo $LC_DNS_A | cut -d ' ' -f 1)
export DNS_AAAA=$(echo $LC_DNS_AAAA | cut -d ' ' -f 1)

(envsubst '$BASE_DOMAIN$DNS_BIND_ADDRESS$DNS_SERIAL$DNS_A$DNS_AAAA' < $LC_BASE_DIR/dns.zone.template) > $LC_BASE_DIR/dns.zone

# Add A and AAA records for all server IP addresses
for ipv4 in $LC_DNS_A; do
  echo "server-multiple-addresses IN A $ipv4" >> $LC_BASE_DIR/dns.zone
done
for ipv6 in $LC_DNS_AAAA; do
  echo "server-multiple-addresses IN AAAA $ipv6" >> $LC_BASE_DIR/dns.zone
done

# Start DNS server
$DOCKER_COMPOSE_CMD up -d --build dns
