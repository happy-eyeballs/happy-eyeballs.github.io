#!/bin/bash

sudo tc qdisc add dev $LC_NETEM_INTERFACE root handle 1: prio

# the delay should be greater than (number of RR) * (max client attempt delay)
sudo tc qdisc add dev $LC_NETEM_INTERFACE parent 1:3 netem delay 30000ms 0ms

sudo tc filter add dev $LC_NETEM_INTERFACE protocol ipv6 parent 1:0 prio 1 u32 match ip6 sport 80 0xffff flowid 1:3
sudo tc filter add dev $LC_NETEM_INTERFACE protocol ip parent 1:0 prio 2 u32 match ip sport 80 0xffff flowid 1:3
