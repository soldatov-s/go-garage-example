package testv1

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage-example/internal/cfg"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
)

const (
	DomainName = "testv1"
)

type empty struct{}

type TestV1 struct {
	log   zerolog.Logger
	ctx   context.Context
	db    *pq.Enity
	cache *redis.Enity
	msgs  *rabbitmq.Enity
}

func Registrate(ctx context.Context) (context.Context, error) {
	t := &TestV1{
		ctx: ctx,
		log: logger.GetPackageLogger(ctx, empty{}),
	}
	var err error
	if t.db, err = pq.GetEnityTypeCast(ctx, cfg.DBName); err != nil {
		return nil, err
	}

	if t.cache, err = redis.GetEnityTypeCast(ctx, cfg.CacheName); err != nil {
		return nil, err
	}

	if t.msgs, err = rabbitmq.GetEnityTypeCast(ctx, cfg.MsgsName); err != nil {
		return nil, err
	}

	publicV1, err := echo.GetAPIVersionGroup(ctx, cfg.PublicHTTP, cfg.V1)
	if err != nil {
		return nil, err
	}

	grPublic := publicV1.Group
	grPublic.Use(echo.HydrationLogger(&t.log))
	grPublic.POST("/test/:id", echo.Handler(t.testPostToCacheHandler))

	privateV1, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.V1)
	if err != nil {
		return nil, err
	}

	grProtect := privateV1.Group
	grProtect.Use(echo.HydrationLogger(&t.log))
	grProtect.GET("/test/:id", echo.Handler(t.testGetHandler))
	grProtect.POST("/test", echo.Handler(t.testPostHandler))
	grProtect.DELETE("/test/:id", echo.Handler(t.testDeleteHandler))

	return domains.RegistrateByName(ctx, DomainName, t), nil
}

func Get(ctx context.Context) (*TestV1, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(*TestV1); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
