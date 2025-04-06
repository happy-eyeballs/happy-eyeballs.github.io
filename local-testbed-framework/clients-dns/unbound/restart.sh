#!/bin/bash

DOCKER_COMPOSE_CMD="docker compose -f $LC_CLIENT_DIR/docker-compose.yml"

$DOCKER_COMPOSE_CMD kill
$DOCKER_COMPOSE_CMD rm -f
docker volume prune -f
$LC_CLIENT_DIR/start.sh
