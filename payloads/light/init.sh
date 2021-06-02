#!/bin/bash

# wait for other nodes to start
sleep 25

# export the public ips of the other nodes
source /root/light/public_ipv4s.sh

# init ipfs
/root/light/das init

# install the hydra-booster node as a bootstrap node
/root/light/das add-hydra "$dht1sgp1":7779 /root/ipfs/config

# get the *latest* data availability header and use it to sample via IPFS. Do this 10 times
/root/light/das sample "$validator1nyc3":26657 10