package apiv1

import (
	"context"

	"github.com/pkg/errors"
	rediscache "github.com/soldatov-s/go-garage/providers/redis/cache"
)

type CacheDeps struct {
	Cache *rediscache.Cache
}

type Cache struct {
	cache *rediscache.Cache
}

func NewCache(deps *CacheDeps) *Cache {
	return &Cache{
		cache: deps.Cache,
	}
}

func (c *Cache) Get(ctx context.Context, key string, value *string) error {
	if err := c.cache.Get(ctx, key, value); err != nil {
		return errors.Wrap(err, "get from cache")
	}

	return nil
}

func (c *Cache) Set(ctx context.Context, key string, value *Enity) error {
	if err := c.cache.Set(ctx, key, value); err != nil {
		return errors.Wrap(err, "set to cache")
	}

	return nil
}
