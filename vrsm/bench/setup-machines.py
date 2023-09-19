#!/usr/bin/env python3
from os import system as do
from subprocess import Popen

procs = []
for i in range(0, 8):
    # this saves the key in known_hosts without asking for the user to
    # interactively type "yes"
    do(f"ssh -o StrictHostKeyChecking=no node{i} 'echo starting node{i}'")

    do(f"scp cloudlab-setup.sh node{i}:")
    do(f"scp .zshrc-cloudlab node{i}:.zshrc")
    c = f"ssh node{i} 'chmod +x cloudlab-setup.sh && ./cloudlab-setup.sh'"
    procs.append(Popen(c, shell=True))

for p in procs:
   p.wait()

# connect between cloudlab nodes so user is not prompted for "yes" later.
for j in [4]:
    for i in range(0, 8):
        do(f"ssh node{j} \"ssh -o StrictHostKeyChecking=no node{i} 'echo connecting node{j} to node{i}'\"")
