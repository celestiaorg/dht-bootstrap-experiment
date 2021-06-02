#!bin/bash

# wait for the hydra node to start
sleep 10

# export the public ips of the other nodes
source /root/validator/public_ipv4s.sh

# add the hydra-booster node as bootstrap dht node
/root/validator/das add-hydra "$dht1sgp1":7779 /root/.tendermint/ipfs/config