#!/bin/bash

DOCKER_COMPOSE_CMD="docker compose -f $LC_CLIENT_DIR/docker-compose.yml"

$DOCKER_COMPOSE_CMD up -d

sleep 5

dig +time=10 -p 15353 example.com. @127.0.0.1 || bash -c 'set -e; sleep 5; dig +time=10 -p 15353 example.com. @127.0.0.1' || $LC_CLIENT_DIR/restart.sh
