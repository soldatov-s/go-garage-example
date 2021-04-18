package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
)

type Cacher interface {
	Get(ctx context.Context, key string, value *string) error
	Set(ctx context.Context, key string, value *Enity) error
	Ping(ctx context.Context) error
}

type Cache struct {
	cache *redis.Enity
}

func NewCache(cache *redis.Enity) *Cache {
	return &Cache{
		cache: cache,
	}
}

func (c *Cache) Get(ctx context.Context, key string, value *string) error {
	if err := c.cache.Get(key, value); err != nil {
		return errors.Wrap(err, "get from cache")
	}

	return nil
}

func (c *Cache) Set(ctx context.Context, key string, value *Enity) error {
	if err := c.cache.Set(key, value); err != nil {
		return errors.Wrap(err, "set to cache")
	}

	return nil
}

func (c *Cache) Ping(ctx context.Context) error {
	if _, err := c.cache.Conn.Ping(context.Background()).Result(); err != nil {
		return errors.Wrap(err, "ping cache")
	}

	return nil
}
