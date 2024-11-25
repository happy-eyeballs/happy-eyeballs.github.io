#!/bin/bash

SCRIPTDIR=$(dirname "$(readlink -f "$0")")

DOCKER_COMPOSE_CMD="docker-compose -f ${SCRIPTDIR}/docker-compose.yml"

$DOCKER_COMPOSE_CMD down
