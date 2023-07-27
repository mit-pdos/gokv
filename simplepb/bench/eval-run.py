#!/usr/bin/env python3
from os import system as do
import os

os.chdir(os.path.expanduser("~/gokv/simplepb"))
do("python -m bench.experiments.e2")
do("python -m bench.experiments.e1")
do("python -m bench.experiments.e3")
