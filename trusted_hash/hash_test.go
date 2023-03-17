package trusted_hash

import "testing"

func TestHash(t *testing.T) {
	hash1 := Hash([]byte{1, 2, 3})
	hash2 := Hash([]byte{3, 2, 1})
	if len(hash1) != 512/8 {
		t.Errorf("incorrect length for hash: %d", len(hash1))
	}
	if hash1 == hash2 {
		t.Errorf("hash collision: %v", []byte(hash1))
	}
}
