package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/logger"
)

type Cacher interface {
	Get(key string, value *string) error
	Set(key string, value *Enity) error
	Ping() error
}

type Cache struct {
	cache *redis.Enity
	log   zerolog.Logger
}

func NewCache(ctx context.Context, cacheName string) (*Cache, error) {
	c := &Cache{}

	var err error
	if c.cache, err = redis.GetEnityTypeCast(ctx, cacheName); err != nil {
		return nil, errors.Wrap(err, "failed to get redis enity")
	}

	c.log = logger.GetPackageLogger(ctx, empty{})

	return c, nil
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
