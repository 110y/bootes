package cache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

type Cache struct {
	cache.SnapshotCache
}

func NewCache() *Cache {
	sc := cache.NewSnapshotCache(false, cache.IDHash{}, nil)
	return &Cache{SnapshotCache: sc}
}
