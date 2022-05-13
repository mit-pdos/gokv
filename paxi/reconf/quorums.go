package reconf

import (
	"github.com/mit-pdos/gokv/grove_ffi"
)

type Config struct {
	members     []grove_ffi.Address
	nextMembers []grove_ffi.Address
}

// Returns some integer i with the property that
// there exists W such that W contains a majority of members and of nextMembers,
// and every node n in W has indices[n] >= i.
// Even more precisely, it returns the largest such i.
func GetHighestIndexOfQuorum(config *Config, indices map[grove_ffi.Address]uint64) uint64 {
	// Will fill orderedIndices with indices of config.members, keeping only the
	// smallest ceil(n/2) values.
	var orderedIndices = make([]uint64, (len(config.members)+1)/2)
	for _, m := range config.members {
		indexToInsert := indices[m]

		// search for where indexToInsert would belong
		for i := range orderedIndices {
			if orderedIndices[i] > indexToInsert {
				// insert indexToInsert at position i, and move everything else
				// to the right
				for j := uint64(i); j < uint64(len(orderedIndices))-1; j += 1 {
					orderedIndices[i+1] = orderedIndices[i]
				}
			}
		}
	}
	ret := orderedIndices[len(config.members)-1]
	if len(config.nextMembers) == 0 {
		return ret
	}
	return 0
}

// Returns true iff w is a (write) quorum for the config `config`.
func IsQuorum(config *Config, w map[grove_ffi.Address]bool) bool {
	var num uint64
	for _, member := range config.members {
		if w[member] {
			num += 1
		}
	}
	if 2*num <= uint64(len(config.members)) {
		return false
	}
	if len(config.nextMembers) == 0 {
		return true
	}

	num = 0
	for _, member := range config.nextMembers {
		if w[member] {
			num += 1
		}
	}
	if 2*num <= uint64(len(config.nextMembers)) {
		return false
	}
	return true
}

func (c *Config) ForEachMember(f func(grove_ffi.Address)) {
	for _, member := range c.members {
		f(member)
	}
	for _, member := range c.nextMembers {
		f(member)
	}
}
