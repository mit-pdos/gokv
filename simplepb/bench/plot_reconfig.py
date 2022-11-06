#!/usr/bin/env python3

import matplotlib.pyplot as plt
import csv

def plot(filename):
    xs = []
    ys = []
    with open('data/' + filename, newline='') as f:
        rows = csv.reader(f)
        for row in rows:
            print(row)
            xs.append(float(row[0]))
            ys.append(int(row[1]))

    f, ax = plt.subplots(1)

    f.set_size_inches(8, 4)
    font = {'fontname':'Computer Modern'}
    ax.plot(xs, ys, marker='', color="blue", label=None)
    ax.set_ylim(bottom=0)
    ax.set_xlim(left=0, right=70)
    ax.set_ylabel('Instantaneous throughput (puts/sec)', **font)
    ax.set_xlabel('Time (s)', **font)

plot('pb_reconfig.dat')
plt.savefig("reconfig_graph.svg")
plt.show()
