package main

import "time"

var w = 10 * time.Second
var e = 60 * time.Second

// TODO: these experiments aren't really independent b/c GoKV will write the
// whole KV store, so probably should add a Clear() RPC or something
func genSet(size int, rate float32) []Experiment {
	return []Experiment{
		&GooseKVPutThroughputExperiment{
			Rate:           rate,
			NumKeys:        1000000,
			WarmupTime:     w,
			ExperimentTime: e,
			ValueGenerator: &RandFixedSizeValueGenerator{
				Size: size,
			},
		},
		&RedisPutThroughputExperiment{
			Rate:           rate,
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
		e = append(e, genSet(128, float32(i)*20.0)...)
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
	&GooseKVPutThroughputExperiment{
		Rate:           1000000,
		NumKeys:        1000,
		WarmupTime:     w,
		ExperimentTime: e,
		ValueGenerator: &RandFixedSizeValueGenerator{
			Size: 128,
		},
	},
}
