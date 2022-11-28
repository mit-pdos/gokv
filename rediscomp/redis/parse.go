package redis

import (
	"bytes"
	"log"
)

func ParseSetCommand(data []byte) ([]byte, []byte) {
	expectedPrefix := []byte("*3\r\n$3\r\nSET\r\n$")
	if !bytes.HasPrefix(data, expectedPrefix) {
		log.Fatalf("unexpected command; got %s", string(data))
	}
	data = data[len(expectedPrefix):]

	keyLen := 0

	// get size of key
	i := 0
	for data[i] != '\r' {
		keyLen = (keyLen * 10) + int(data[i]-'0')
		i++
	}

	if !(data[i+1] == '\n') {
		panic("expected LF")
	}
	data = data[i+2:]

	// now get the key
	if len(data) < keyLen+3 { // + 3 for \r\n$
		panic("incomplete SET command")
	}

	key := data[:keyLen]
	data = data[keyLen+3:]

	// get size of value
	i = 0
	valLen := 0
	for data[i] != '\r' {
		valLen = (valLen * 10) + int(data[i]-'0')
		i++
	}

	if !(data[i+1] == '\n') {
		panic("expected LF")
	}
	data = data[i+2:]

	val := data[:valLen]

	if !bytes.Equal(data[valLen:], []byte("\r\n")) {
		log.Fatalf("unexpected data after end of SET command; have %s", data[valLen:])
	}

	return key, val
}
