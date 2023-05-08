#!/usr/bin/env python3
from os import system as do

do("ssh-keygen -C 'upamanyu' -f /tmp/id_rsa -N ''")
for i in range(0, 8):
    do(f"scp cloudlab-setup.sh node{i}:")
    do(f"scp ~/.zshrc-cloudlab node{i}:.zshrc")
    do(f"ssh upamanyu@node{i} 'chmod +x cloudlab-setup.sh && ./cloudlab-setup.sh'")

    # set up SSH
    do(f"scp /tmp/id_rsa node{i}:.ssh/")
    do(f"ssh upamanyu@node{i} 'ssh-keygen -y -f ~/.ssh/id_rsa >> ~/.ssh/authorized_keys'")
