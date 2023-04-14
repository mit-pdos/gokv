#!/usr/bin/env bash

export PATH=$PATH:/usr/local/go/bin

# download and set up Go, if not already there
if ! [[ $(which go) ]]; then
    cd /tmp
    wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
fi

# set shell:
sudo chsh -s /bin/zsh upamanyu

# turn off hyperthreading
echo off | sudo tee /sys/devices/system/cpu/smt/control

# install numactl
sudo apt update
yes | sudo apt install numactl cmake
