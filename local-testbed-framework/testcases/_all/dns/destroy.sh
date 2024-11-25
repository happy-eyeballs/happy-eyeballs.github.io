#!/bin/bash

DOCKER_COMPOSE_CMD="docker compose -f $LC_BASE_DIR/docker-compose.yml"

$DOCKER_COMPOSE_CMD down
