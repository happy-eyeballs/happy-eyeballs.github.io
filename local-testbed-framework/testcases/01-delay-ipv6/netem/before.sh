#!/bin/bash

sudo tc qdisc add dev $LC_NETEM_INTERFACE root handle 1: prio

sudo tc qdisc add dev $LC_NETEM_INTERFACE parent 1:3 netem delay "${LC_NETEM_DELAY}ms" 0ms

sudo tc filter add dev $LC_NETEM_INTERFACE protocol ipv6 parent 1:0 prio 1 u32 match ip6 sport 80 0xffff flowid 1:3
