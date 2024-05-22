package shard

import "sync"

type Shard struct {
	Data map[string]interface{}
	Mu sync.RWMutex
}

type ShardList struct {
	Shards []*Shard
}