package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
)

const DomainName = "testv1"

type Config struct {
	DBName      string
	CacheName   string
	MsgsName    string
	PublicHTTP  string
	PrivateHTTP string
	Version     string
}

type Object struct {
	repository RepositoryGateway
	handler    HandlerGateway
	mess       Messenger
	cache      Cacher
}

func NewObject(ctx context.Context, cfg *Config) (*Object, error) {
	o := &Object{}

	if enity, err := pq.GetEnityTypeCast(ctx, cfg.DBName); err == nil {
		o.repository = NewRepo(enity)
	} else {
		return nil, errors.Wrap(err, "failed to get pq enity")
	}

	if enity, err := redis.GetEnityTypeCast(ctx, cfg.CacheName); err == nil {
		o.cache = NewCache(enity)
	} else {
		return nil, errors.Wrap(err, "failed to get redis enity")
	}

	if enity, err := rabbitmq.GetEnityTypeCast(ctx, cfg.MsgsName); err == nil {
		o.mess = NewMess(&MessDeps{Msgs: enity, Repository: o.repository, Cache: o.cache})
	} else {
		return nil, errors.Wrap(err, "failed to get rabbitmq enity")
	}

	o.handler = NewHandler(&HandlerDeps{Repository: o.repository, Cache: o.cache})

	publicGroup, err := echo.GetAPIVersionGroup(ctx, cfg.PublicHTTP, cfg.Version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get public http enity")
	}

	privateGroup, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.Version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get private http enity")
	}

	o.handler.SetRoutes(publicGroup, privateGroup)

	return o, nil
}

func (o *Object) GetRepository() RepositoryGateway {
	return o.repository
}

func (o *Object) GetCache() Cacher {
	return o.cache
}

func (o *Object) GetHandler() HandlerGateway {
	return o.handler
}

func (o *Object) GetMess() Messenger {
	return o.mess
}

type ObjectInterface interface {
	GetRepository() RepositoryGateway
	GetHandler() HandlerGateway
	GetMess() Messenger
	GetCache() Cacher
}

func Registrate(ctx context.Context, cfg *Config) (context.Context, error) {
	o, err := NewObject(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create interface")
	}

	return domains.RegistrateByName(ctx, DomainName, ObjectInterface(o)), nil
}

func Get(ctx context.Context) (ObjectInterface, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(ObjectInterface); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
