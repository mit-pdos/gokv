package kv

type Kv interface {
	Put(key, value string)
	Get(key string) string
}

type KvCput interface {
	Kv
	ConditionalPut(key, expect, value string) string
}
