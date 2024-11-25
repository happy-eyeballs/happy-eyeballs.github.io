#!/bin/bash

set -x
set -e


INTERFACE=enp0s13f0u1u4

sudo ip a add dev $INTERFACE 10.0.2.1/16
sudo ip a add dev $INTERFACE 10.0.2.2/16
sudo ip a add dev $INTERFACE 10.0.2.3/16
sudo ip a add dev $INTERFACE 10.0.2.4/16
sudo ip a add dev $INTERFACE 10.0.2.5/16
sudo ip a add dev $INTERFACE 10.0.2.6/16
sudo ip a add dev $INTERFACE 10.0.2.7/16
sudo ip a add dev $INTERFACE 10.0.2.8/16
sudo ip a add dev $INTERFACE 10.0.2.9/16
sudo ip a add dev $INTERFACE 10.0.2.10/16

sudo ip a add dev $INTERFACE 3000::2:1/64
sudo ip a add dev $INTERFACE 3000::2:2/64
sudo ip a add dev $INTERFACE 3000::2:3/64
sudo ip a add dev $INTERFACE 3000::2:4/64
sudo ip a add dev $INTERFACE 3000::2:5/64
sudo ip a add dev $INTERFACE 3000::2:6/64
sudo ip a add dev $INTERFACE 3000::2:7/64
sudo ip a add dev $INTERFACE 3000::2:8/64
sudo ip a add dev $INTERFACE 3000::2:9/64
sudo ip a add dev $INTERFACE 3000::2:10/64
