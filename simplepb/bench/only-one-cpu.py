#!/usr/bin/env python3

from os import system as do
for i in [1,3,5,7,9,11,13,15]:
    do(f"echo 0 | sudo tee /sys/devices/system/cpu/cpu{i}/online")
