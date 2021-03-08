package main

import "time"

var w = 2 * time.Second
var e = 10 * time.Second

// TODO: these experiments aren't really independent b/c GoKV will write the
// whole KV store, so probably should add a Clear() RPC or something
func genSet(size, numClients int) []Experiment {
	return []Experiment{
		&PutThroughputExperiment{
			NumClients:     numClients,
			NumKeys:        1000000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
		&RedisPutThroughputExperiment{
			NumClients:     numClients,
			NumKeys:        1000000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
	}
}

func genLT() []Experiment {
	var e []Experiment
	for i := 0; i < 30; i++ {
		e = append(e, genSet(128, i*20)...)
	}

	// e = append(e, genSet(128, 1)...)
	// e = append(e, genSet(128, 2)...)
	// e = append(e, genSet(128, 4)...)
	// e = append(e, genSet(128, 8)...)
	// e = append(e, genSet(128, 16)...)
	// e = append(e, genSet(128, 32)...)
	// e = append(e, genSet(128, 64)...)
	// e = append(e, genSet(128, 128)...)
	// e = append(e, genSet(128, 256)...)
	// e = append(e, genSet(128, 512)...)
	// e = append(e, genSet(128, 1024)...)
	return e
}

// var experiments = genLT()
var experiments = []Experiment{
	&PutThroughputExperiment{
		NumClients:     1000,
		NumKeys:        100000,
		WarmupTime:     w,
		ExperimentTime: e,
		ValueGenerator: &RandFixedSizeValueGenerator{
			Size: 128,
		},
	},
}
