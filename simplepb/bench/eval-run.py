#!/usr/bin/env python3
from os import system as do
import os

os.chdir(os.path.expanduser("~/gokv/simplepb"))
do("python -m bench.experiments.e2")
print("Done with reconfiguration experiment")

do("python -m bench.experiments.e1")
print("Done with redis vs grove experiments")

do("python -m bench.experiments.e3")
print("Done with multi-server scaling experiment")
