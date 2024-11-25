#!/bin/bash

DOCKER_COMPOSE_CMD="docker compose -f $LC_CLIENT_DIR/docker-compose.yml"

$DOCKER_COMPOSE_CMD restart
