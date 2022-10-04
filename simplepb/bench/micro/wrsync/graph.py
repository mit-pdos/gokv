#!/usr/bin/env python3

import matplotlib.pyplot as plt
import csv

xs = []
ys = []
with open('data.csv', newline='') as csvfile:
    for row in csv.reader(csvfile, delimiter=','):
        xs.append(int(row[0]))
        ys.append(1/float(row[1]))


plt.plot(xs, ys)
plt.ylabel('sec/write')
plt.xlabel('size of write in bytes')
plt.show()
