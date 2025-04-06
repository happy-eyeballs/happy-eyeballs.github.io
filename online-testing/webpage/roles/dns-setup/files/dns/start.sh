#!/bin/bash

SCRIPTDIR=$(dirname "$(readlink -f "$0")")

${SCRIPTDIR}/destroy.sh

DOCKER_COMPOSE_CMD="docker-compose -f ${SCRIPTDIR}/docker-compose.yml"

# Start DNS server
$DOCKER_COMPOSE_CMD up -d --build dns
