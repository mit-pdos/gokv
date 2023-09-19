#!/usr/bin/env python3

from os import system as do
for i in range(16):
    do(f"echo 1 | sudo tee /sys/devices/system/cpu/cpu{i}/online")
