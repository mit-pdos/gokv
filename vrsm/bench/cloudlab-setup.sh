#!/usr/bin/env bash

# Set up SSH keys to ssh between nodes
/usr/bin/geni-get key > ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa
ssh-keygen -y -f ~/.ssh/id_rsa > ~/.ssh/id_rsa.pub
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod 644 ~/.ssh/authorized_keys

export PATH=$PATH:/usr/local/go/bin

# download and set up Go, if not already there
if ! [[ $(which go) ]]; then
    cd /tmp
    rm go1.19.3.linux-amd64.tar.gz*
    wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
fi

# set shell:
sudo chsh -s /bin/zsh $(whoami)

# turn off hyperthreading
echo off | sudo tee /sys/devices/system/cpu/smt/control

# set intel_pstate driver to use performance governor; note that this is not the
# same as actually setting the CPU frequency to a fixed amount, which does not
# seem easily doable on cloudlab.
echo performance | sudo tee /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor

# install numactl
sudo apt update
yes | sudo apt install numactl cmake pip texlive texlive-latex-extra
pip install numpy scipy
