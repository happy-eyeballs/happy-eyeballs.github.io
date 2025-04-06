#!/bin/bash

sudo tc qdisc del dev $LC_NETEM_INTERFACE parent 1:3
sudo tc qdisc del dev $LC_NETEM_INTERFACE root