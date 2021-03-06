package main

import "time"

var w = 10 * time.Second
var e = 20 * time.Second

// TODO: these experiments aren't really independent b/c GoKV will write the
// whole KV store, so probably should add a Clear() RPC or something
func genSet(size int) []Experiment {
	return []Experiment{
		&PutThroughputExperiment{
			NumClients:     100,
			NumKeys:        1000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
		&RedisPutThroughputExperiment{
			NumClients:     100,
			NumKeys:        1000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
		&PutThroughputExperiment{
			NumClients:     100,
			NumKeys:        1000000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
		&RedisPutThroughputExperiment{
			NumClients:     100,
			NumKeys:        1000000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
	}
}

var experiments = append(genSet(128), genSet(4096)...)
