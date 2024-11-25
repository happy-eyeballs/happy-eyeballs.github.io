#!/bin/bash

export PATH=$PATH:/usr/local/bin

DOCKER_COMPOSE_CMD="docker compose -f $LC_BASE_DIR/docker-compose.yml"

# create necessary directories for an easy download of artifacts
mkdir -p $LC_BASE_DIR/pcaps

# start capture
if [ "$LC_USE_DOCKER_TCPDUMP" == "true" ]
then
  $DOCKER_COMPOSE_CMD up -d tcpdump
else
  tmux new-session -d -s tcpdump_session "sudo tcpdump -U -i $LC_TCPDUMP_INTERFACE -w $LC_BASE_DIR/pcaps/capture.pcap '(port 80 or port 53) or (udp and port 44444 and dst $LC_TCPDUMP_SEPARATOR_PACKET_DESTINATION_ADDRESS)'";
fi

# wait a few seconds to make sure tcpdump is capturing our packets
sleep 2
