package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
)

const (
	domainName = "testv1"
)

type Config struct {
	DBName      string
	CacheName   string
	MsgsName    string
	PublicHTTP  string
	PrivateHTTP string
	Version     string
}

type Object struct {
	Repo  Repository
	API   APIInterface
	Mess  Messenger
	Cache Cacher
}

func NewObject(ctx context.Context, cfg *Config) (*Object, error) {
	o := &Object{}
	var err error
	if o.Repo, err = NewRepository(ctx, cfg.DBName); err != nil {
		return nil, errors.Wrap(err, "create repository")
	}

	if o.Cache, err = NewCache(ctx, cfg.CacheName); err != nil {
		return nil, errors.Wrap(err, "create cache")
	}

	if o.Mess, err = NewMess(ctx, cfg.MsgsName, o.Repo, o.Cache); err != nil {
		return nil, errors.Wrap(err, "create messager")
	}

	o.API = NewAPI(o.Repo, o.Cache)

	publicV1, err := echo.GetAPIVersionGroup(ctx, cfg.PublicHTTP, cfg.Version)
	if err != nil {
		return nil, errors.Wrap(err, "get version group")
	}

	grPublic := publicV1.Group
	log := logger.GetPackageLogger(ctx, empty{})
	grPublic.Use(echo.HydrationLogger(&log))
	grPublic.POST("/test/:id", echo.Handler(o.API.PostToCacheHandler))

	privateV1, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.Version)
	if err != nil {
		return nil, errors.Wrap(err, "get version group")
	}

	grProtect := privateV1.Group
	grProtect.Use(echo.HydrationLogger(&log))
	grProtect.GET("/test/:id", echo.Handler(o.API.GetHandler))
	grProtect.POST("/test", echo.Handler(o.API.PostHandler))
	grProtect.DELETE("/test/:id", echo.Handler(o.API.DeleteHandler))

	return o, nil
}

func Registrate(ctx context.Context, cfg *Config) (context.Context, error) {
	i, err := NewObject(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create interface")
	}

	return domains.RegistrateByName(ctx, domainName, i), nil
}

func Get(ctx context.Context) (*Object, error) {
	if v, ok := domains.GetByName(ctx, domainName).(*Object); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
