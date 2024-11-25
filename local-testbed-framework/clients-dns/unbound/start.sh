#!/bin/bash

export
(envsubst '$LC_KNOT_VERSION' < $LC_BASE_DIR/build/Dockerfile.tmp) > $LC_BASE_DIR/build/Dockerfile

DOCKER_COMPOSE_CMD="docker compose -f $LC_CLIENT_DIR/docker-compose.yml"

$DOCKER_COMPOSE_CMD up -d

sleep 5

dig -p 15353 internal.example.com. @127.0.0.1 soa || bash -c 'set -e; sleep 5; dig -p 15353 internal.example.com. @127.0.0.1 soa' || $LC_CLIENT_DIR/restart.sh
