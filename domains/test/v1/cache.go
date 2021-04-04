package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
)

type Cacher interface {
	Get(key string, value *string) error
	Set(key string, value *Enity) error
	Ping() error
}

type Cache struct {
	cache *redis.Enity
	log   *zerolog.Logger
}

var _ Cacher = new(Cache)

func NewCache(log *zerolog.Logger, cache *redis.Enity) *Cache {
	return &Cache{
		log:   log,
		cache: cache,
	}
}

func (c *Cache) Get(key string, value *string) error {
	if err := c.cache.Get(key, value); err != nil {
		c.log.Debug().Msgf("not find key %q in cache", key)
		return errors.Wrap(err, "get from cache")
	}

	return nil
}

func (c *Cache) Set(key string, value *Enity) error {
	if err := c.cache.Set(key, value); err != nil {
		c.log.Err(err).Msgf("failed to set data in cache, key %q, data %+v", key, value)
		return errors.Wrap(err, "set to cache")
	}

	return nil
}

func (c *Cache) Ping() error {
	if _, err := c.cache.Conn.Ping(context.Background()).Result(); err != nil {
		c.log.Debug().Err(err).Msg("ping cache")
		return errors.Wrap(err, "ping cache")
	}

	return nil
}
