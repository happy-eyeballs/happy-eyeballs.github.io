#!/bin/bash

export PATH=$PATH:/usr/local/bin

DOCKER_COMPOSE_CMD="docker compose -f $LC_BASE_DIR/docker-compose.yml"

if [ "$LC_USE_DOCKER_TCPDUMP" == "true" ]
then
  $DOCKER_COMPOSE_CMD down
fi
