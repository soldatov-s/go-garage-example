package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
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

type ObjectInterface interface {
	GetRepo() Repository
	GetAPI() APIInterface
	GetMess() Messenger
	GetCache() Cacher
}

type Object struct {
	repo  Repository
	api   APIInterface
	mess  Messenger
	cache Cacher
}

var _ ObjectInterface = new(Object)

// Private type, used for configure logger
type empty struct{}

func NewObject(ctx context.Context, cfg *Config) (*Object, error) {
	o := &Object{}

	log := logger.GetPackageLogger(ctx, empty{})

	if enity, err := pq.GetEnityTypeCast(ctx, cfg.DBName); err == nil {
		o.repo = NewRepo(&log, enity)
	} else {
		return nil, errors.Wrap(err, "failed to get pq enity")
	}

	if enity, err := redis.GetEnityTypeCast(ctx, cfg.CacheName); err == nil {
		o.cache = NewCache(&log, enity)
	} else {
		return nil, errors.Wrap(err, "failed to get redis enity")
	}

	if enity, err := rabbitmq.GetEnityTypeCast(ctx, cfg.MsgsName); err == nil {
		o.mess = NewMess(&log, enity, o.repo, o.cache)
	} else {
		return nil, errors.Wrap(err, "failed to get rabbitmq enity")
	}

	o.api = NewAPI(o.repo, o.cache)

	publicHTTP, err := echo.GetEnityTypeCast(ctx, cfg.PublicHTTP)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get public http enity")
	}

	privateHTTP, err := echo.GetEnityTypeCast(ctx, cfg.PrivateHTTP)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get private http enity")
	}

	if err := o.api.SetRoutes(&log, publicHTTP, privateHTTP, cfg.Version); err != nil {
		return nil, errors.Wrap(err, "set routes")
	}

	return o, nil
}

func (o *Object) GetRepo() Repository {
	return o.repo
}

func (o *Object) GetCache() Cacher {
	return o.cache
}

func (o *Object) GetAPI() APIInterface {
	return o.api
}

func (o *Object) GetMess() Messenger {
	return o.mess
}

func Registrate(ctx context.Context, cfg *Config) (context.Context, error) {
	o, err := NewObject(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create interface")
	}

	return domains.RegistrateByName(ctx, domainName, ObjectInterface(o)), nil
}

func Get(ctx context.Context) (ObjectInterface, error) {
	if v, ok := domains.GetByName(ctx, domainName).(ObjectInterface); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
