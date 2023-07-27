#!/usr/bin/env python3
from os import system as do

do("ssh-keygen -C '$(whoami)' -f /tmp/id_rsa -N ''")
for i in range(0, 8):
    # this saves the key in known_hosts without asking for the user to
    # interactively type "yes"
    do(f"ssh -o StrictHostKeyChecking=no node{i} 'echo starting node{i}'")

    do(f"scp cloudlab-setup.sh node{i}:")
    do(f"scp .zshrc-cloudlab node{i}:.zshrc")
    do(f"ssh node{i} 'chmod +x cloudlab-setup.sh && ./cloudlab-setup.sh'")

    # set up SSH
    do(f"scp /tmp/id_rsa node{i}:.ssh/")
    do(f"ssh node{i} 'ssh-keygen -y -f ~/.ssh/id_rsa >> ~/.ssh/authorized_keys'")

# connect between cloudlab nodes so user is not prompted for "yes" later.
for j in [4]:
    for i in range(0, 8):
        do(f"ssh node{j} \"ssh -o StrictHostKeyChecking=no node{i} 'echo connecting node{j} to node{i}'\"")
