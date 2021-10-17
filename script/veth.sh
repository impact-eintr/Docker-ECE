#!/usr/bin/env bash
sudo ip netns add ns1
sudo ip netns add ns2
sudo ip link add veth0 type veth peer name veth1
sudo ip link set veth0 netns ns1
sudo ip link set veth1 netns ns2
sudo ip netns exec ns1 ifconfig veth0 172.18.0.2/24 up
sudo ip netns exec ns2 ifconfig veth1 172.18.0.3/24 up
sudo ip netns exec ns1 route add default dev veth0
sudo ip netns exec ns2 route add default dev veth1
sudo ip netns exec ns1 ping -c 1 172.18.0.3
