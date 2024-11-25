#!/bin/bash

sudo tc qdisc add dev $LC_NETEM_INTERFACE root handle 1: prio

sudo tc class add dev $LC_NETEM_INTERFACE parent 1: classid 1:3 htb rate 10gbit

sudo tc qdisc add dev $LC_NETEM_INTERFACE parent 1:3 netem delay "${LC_NETEM_DELAY}ms"

DNS_AAAA=$(echo $LC_DNS_AAAA | cut -d ' ' -f 2)
sudo tc filter add dev $LC_NETEM_INTERFACE protocol ipv6 parent 1: prio 1 u32 match ip6 src $DNS_AAAA/128 match ip6 protocol 17 0xff flowid  1:3
