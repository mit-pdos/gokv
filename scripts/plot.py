#!/usr/bin/env python3

import matplotlib.pyplot as plt
import csv
import sys

# TODO: deal with warmup
def get_latencies(fnames):
    x = []
    ys = []
    lastStart = -1
    title = ''
    with open(fname,'r') as csvfile:
        reader = csv.reader(csvfile, delimiter=',')
        title = next(reader, None)[0]
        for row in reader:
            i = row[0]
        # This might be a bit fragile
        if i.endswith("Beg"):
            lastStart = int(row[1])
        else:
            ys.append( (int(row[1]) - lastStart) / 1e6 )
    return title, ys

def plot_latencies(fname):
    title, ys = get_latencies(fname)

    plt.plot(ys, label='')
    plt.xlabel('Operation')
    plt.ylabel('Latency in microseconds')
    plt.title(title)
    plt.legend()
    plt.show()

def plot_latencythruput(fnames):
    # want to compute latency and thruputs for each file
    ls = []
    ths = []
    for fname in fnames:
        with open(fname,'r') as csvfile:
            lasty = -1
            reader = csv.reader(csvfile, delimiter=',')
            title = next(reader, None)[0]
            lastStart = -1
            begin = 2**64
            end = 0
            ys = []
            for row in reader:
                y = int(row[1])
                if y < lasty:
                    lastStart = -1
                lasty = y

                begin = min(begin, y)
                end = max(end, y)

                i = row[0]
                if i.endswith("Beg"):
                    lastStart = y
                elif lastStart != -1:
                    if y < lastStart:
                        print("Panic")
                        print(y, lastStart)
                    ys.append( (y - lastStart) / 1e3 )

            if len(ys) == 0:
                continue
            th = 1e9 * len(ys)/float(end - begin) # in op/sec
            avg_latency = sum(ys)/len(ys) # in us
            print(th, avg_latency)
            ls.append(avg_latency)
            ths.append(th)

    plt.plot(ls, ths)
    plt.xlabel('Throughput, ops/sec')
    plt.ylabel('Latency in microseconds')
    plt.title(title)
    plt.legend()
    plt.show()

# plot_timeseries(sys.argv[1])
plot_latencythruput(sys.argv[1:])
