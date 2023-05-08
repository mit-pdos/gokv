#!/usr/bin/env python3
from os import system as do
import os

os.chdir("../../../redis/")
do("cd redis/src && make")
do("rm -rf redisraft/build && mkdir redisraft/build")
os.chdir("redisraft/build")
do("cmake .. && make")
