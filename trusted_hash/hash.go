package trusted_hash

import "crypto/sha512"

func Hash(data []byte) string {
	hasher := sha512.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	return string(hash)
}
