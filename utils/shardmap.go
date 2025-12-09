// Implement a shard map to store key-value pairs

package utils

import (
	"hash/fnv"
	"sync"
)

const (
	ShardCount = 1024
)

type Shard struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

type ShardMap struct {
	shards []*Shard
}

func fnv32(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (sm *ShardMap) getShard(key string) *Shard {
	hash := fnv32(key)
	return sm.shards[hash%ShardCount]
}

func NewShardMap() *ShardMap {
	sm := &ShardMap{
		shards: make([]*Shard, ShardCount),
	}
	for i := 0; i < ShardCount; i++ {
		sm.shards[i] = &Shard{
			data: make(map[string]interface{}),
		}
	}
	return sm
}

func (sm *ShardMap) Set(key string, value interface{}) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.data[key] = value
}

func (sm *ShardMap) Get(key string) interface{} {
	shard := sm.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	return shard.data[key]
}
