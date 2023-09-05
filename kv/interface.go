package kv

type Kv struct {
	Put            func(key, value string)
	Get            func(key string) string
	ConditionalPut func(key, expect, value string) string
}
