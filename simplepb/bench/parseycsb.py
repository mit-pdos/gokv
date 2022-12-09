#!/usr/bin/env python3
import re

def parse_ycsb_output_totalops(output):
    # look for 'Run finished, takes...', then parse the lines for each of the operations
    # output = output[re.search("Run finished, takes .*\n", output).end():] # strip off beginning of output

    # NOTE: sample output from go-ycsb:
    # UPDATE - Takes(s): 12.6, Count: 999999, OPS: 79654.6, Avg(us): 12434, Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000
    patrn = '(?P<opname>.*) - Takes\(s\): (?P<time>.*), Count: (?P<count>.*), OPS: (?P<ops>.*), Avg\(us\): (?P<avg_latency>.*), Min\(us\):.*\n' # Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000'
    ms = re.finditer(patrn, output, flags=re.MULTILINE)
    ops = 0
    time = None
    numMatches = 0
    for m in ms:
        ops += int(m.group('count'))
        time = float(m.group('time'))
        latency = float(m.group('avg_latency'))
        numMatches += 1
    assert numMatches == 1, "expected only one kind of operation, but got " + str(numMatches)
    return (time, ops, latency)
