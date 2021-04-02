package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
)

const (
	DomainName = "testv1"
)

type empty struct{}

type Config struct {
	DBName      string
	CacheName   string
	MsgsName    string
	PublicHTTP  string
	PrivateHTTP string
	Version     string
}

type Interface struct {
	Repo  Repository
	App   AppInterface
	Mess  Messenger
	Cache Cacher
}

func Registrate(ctx context.Context, cfg *Config) (context.Context, error) {
	i := &Interface{}
	var err error
	if i.Repo, err = NewRepository(ctx, &RepoConfig{DBName: cfg.DBName}); err != nil {
		return nil, errors.Wrap(err, "create repository")
	}

	if i.Cache, err = NewCache(ctx, cfg.CacheName); err != nil {
		return nil, errors.Wrap(err, "create cache")
	}

	if i.Mess, err = NewMess(ctx, cfg.MsgsName, i.Repo, i.Cache); err != nil {
		return nil, errors.Wrap(err, "create messager")
	}

	i.App = NewApp(i.Repo, i.Cache)

	publicV1, err := echo.GetAPIVersionGroup(ctx, cfg.PublicHTTP, cfg.Version)
	if err != nil {
		return nil, err
	}

	grPublic := publicV1.Group
	log := logger.GetPackageLogger(ctx, empty{})
	grPublic.Use(echo.HydrationLogger(&log))
	grPublic.POST("/test/:id", echo.Handler(i.App.PostToCacheHandler))

	privateV1, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.Version)
	if err != nil {
		return nil, err
	}

	grProtect := privateV1.Group
	grProtect.Use(echo.HydrationLogger(&log))
	grProtect.GET("/test/:id", echo.Handler(i.App.GetHandler))
	grProtect.POST("/test", echo.Handler(i.App.PostHandler))
	grProtect.DELETE("/test/:id", echo.Handler(i.App.DeleteHandler))

	return domains.RegistrateByName(ctx, DomainName, i), nil
}

func Get(ctx context.Context) (*Interface, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(*Interface); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
