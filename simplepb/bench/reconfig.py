#!/usr/bin/env python3

from os import system as do
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('nreplicas', metavar='nreplicas', type=int,
                    help='number of replicas to configure to')
args = parser.parse_args()

servers = ' '.join([f'10.10.1.{str(i + 1)}:12100' for i in range(args.nreplicas)])
do(f"go run ../cmd/admin -conf 10.10.1.4:12000 reconfig {servers}")
