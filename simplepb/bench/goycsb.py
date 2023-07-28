#!/usr/bin/env python3
from os import path
import re
import json
import time
from .common import *

def parse_ycsb_output(output):
    # look for 'Run finished, takes...', then parse the lines for each of the operations
    # output = output[re.search("Run finished, takes .*\n", output).end():] # strip off beginning of output

    # NOTE: sample output from go-ycsb:
    # UPDATE - Takes(s): 12.6, Count: 999999, OPS: 79654.6, Avg(us): 12434, Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000
    patrn = '(?P<opname>.*)\s+- Takes\(s\): (?P<time>.*), Count: (?P<count>.*), OPS: (?P<ops>.*), Avg\(us\): (?P<avg_latency>.*), Min\(us\):.*\n' # Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000'
    ms = re.finditer(patrn, output, flags=re.MULTILINE)
    a = dict()
    for m in ms:
        a[m.group('opname').strip()] = {'thruput': float(m.group('ops')), 'avg_latency': float(m.group('avg_latency')), 'raw': output[m.span()[0] : m.span()[1]]}
    return a

def goycsb_bench(kvname:str, threads:int, warmuptime:int, runtime:int, valuesize:int, readprop:float, updateprop:float, keys:int, extra_args=[], cooldown=0):
    """
    Returns a dictionary of the form
    { 'UPDATE': {'thruput': 1000, 'avg_latency': 12345', 'raw': 'blah'},...}
    """

    p = start_command([path.join(goycsbdir, './go-ycsb'),
                       'run', kvname,
                       '-P', path.join(simplepbdir, "bench", kvname + '_workload'),
                       '--threads', str(threads),
                       '--target', '-1',
                       '--interval', '200',
                       '-p', 'operationcount=' + str(2**32 - 1),
                       '-p', 'fieldlength=' + str(valuesize),
                       '-p', 'requestdistribution=uniform',
                       '-p', 'readproportion=' + str(readprop),
                       '-p', 'updateproportion=' + str(updateprop),
                       '-p', 'warmuptime=' + str(warmuptime), # TODO: increase warmup
                       '-p', 'recordcount=' + str(keys),
                       ] + extra_args, cwd=goycsbdir)

    if p is None:
        return ''

    to_parse = ""
    optypes_seen = 0
    # XXX: there is an extra line of output for "TOTAL"
    num_optypes = 2 if readprop == 0.0 or updateprop == 0.0 else 3

    for stdout_line in iter(p.stdout.readline, ""):
        if stdout_line.find('Takes(s): {0}.'.format(runtime)) != -1:
          to_parse += stdout_line
          optypes_seen += 1
          if optypes_seen == num_optypes:
              break
    time.sleep(cooldown)
    p.stdout.close()
    p.terminate()
    return parse_ycsb_output(to_parse)

def parse_ycsb_output_totalops(output):
    # look for 'Run finished, takes...', then parse the lines for each of the operations
    # output = output[re.search("Run finished, takes .*\n", output).end():] # strip off beginning of output

    # NOTE: sample output from go-ycsb:
    # UPDATE - Takes(s): 12.6, Count: 999999, OPS: 79654.6, Avg(us): 12434, Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000
    patrn = '(?P<opname>.*) - Takes\(s\): (?P<time>.*), Count: (?P<count>.*), OPS: (?P<ops>.*), Avg\(us\): (?P<avg_latency>.*), Min\(us\):.*\n' # Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000'
    ms = re.finditer(patrn, output, flags=re.MULTILINE)
    a = 0
    time = None
    for m in ms:
        a += int(m.group('count'))
        time = float(m.group('time'))
    return (time, a)

def goycsb_load(kvname:str, threads:int, valuesize:int, keys:int, extra_args=[]):
    run_command([path.join(goycsbdir, './go-ycsb'),
                 'load', kvname,
                 '-P', path.join(simplepbdir, "bench", kvname + '_workload'),
                 '--threads', str(threads),
                 '-p', 'fieldlength=' + str(valuesize),
                 '-p', 'recordcount=' + str(keys),
                 ] + extra_args, cwd=goycsbdir)


# This runs go-ycsb with an increasing number of threads until the throughput is
# saturated, and returns a list of the latency/throughput information for each
# run.
def closed_lt(kvname, config, reset_fn, thread_fn, ycsb_args):
    data = []
    i = 0
    last_good_index = i
    peak_thruput = 0

    while True:
        if i >= last_good_index + 5:
            break
        threads = thread_fn(i)
        if threads is None:
            return data

        # restart and reload every single time
        reset_fn()
        goycsb_load(kvname, 200, config['valuesize'], config['recordcount'], ycsb_args)

        a = goycsb_bench(kvname, threads, config['warmuptime'], config['runtime'],
                         config['valuesize'], config['reads'], config['writes'],
                         config['recordcount'],
                         ycsb_args)

        p = {'service': kvname, 'num_threads': threads, 'lts': a}
        data = data + [p]

        with open(config['outfilename'], 'a+') as outfile:
            outfile.write(json.dumps(p) + '\n')

        thput = sum([ a[op]['thruput'] for op in a ])

        if thput > peak_thruput:
            last_good_index = i
        if thput > peak_thruput:
            peak_thruput = thput

        last_threads = threads
        i = i + 1

    return data
