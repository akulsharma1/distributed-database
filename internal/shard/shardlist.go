package shard

import "crypto/sha1"

func (s *ShardList) GetShardCount() int {
	return len(s.Shards)
}

func (s *ShardList) AddShard() {
	s.Shards = append(s.Shards, CreateShard())
}

func (s *ShardList) AddKeyValuePair(key string, value interface{}) {
	idx := s.CalculateShardIdx(key)

	s.Shards[idx].Put(key, value)
}

func (s *ShardList) CalculateShardIdx(key string) int {
	hash := sha1.Sum([]byte(key))
	return int(hash[0]) % len(s.Shards)
}

func (s *ShardList) ReShard() {
	keyVals := map[string]interface{}{}
	for _, shard := range s.Shards {
		for key, val := range shard.Data {
			keyVals[key] = val
		}
		shard.Data = map[string]interface{}{}
	}

	for key, val := range keyVals {
		s.AddKeyValuePair(key, val)
	}
}