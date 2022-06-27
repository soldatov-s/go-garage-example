package cfg

import (
	"time"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/log"
	"github.com/soldatov-s/go-garage/providers/echo"
	"github.com/soldatov-s/go-garage/providers/pq"
	"github.com/soldatov-s/go-garage/providers/rabbitmq"
	rabbitmqconsum "github.com/soldatov-s/go-garage/providers/rabbitmq/consumer"
	rabbitmqpool "github.com/soldatov-s/go-garage/providers/rabbitmq/pool"
	rabbitmqpub "github.com/soldatov-s/go-garage/providers/rabbitmq/publisher"
	"github.com/soldatov-s/go-garage/providers/redis"
	rediscache "github.com/soldatov-s/go-garage/providers/redis/cache"
	"github.com/soldatov-s/go-garage/x/sqlx/migrations"
	"github.com/vrischmann/envconfig"
)

const (
	DBName    = "dbTest"
	CacheName = "cacheTest"
	MsgsName  = "msgsTest"
	StatsName = "statsTest"

	PublicHTTP  = "public"
	PrivateHTTP = "private"
	V1          = "1"
)

type Config struct {
	Logger      *log.Config
	DB          *pq.Config
	PrivateHTTP *echo.Config
	RabbitMQ    *rabbitmq.Config
	Consumer    *rabbitmqconsum.Config
	Publisher   *rabbitmqpub.Config
	Redis       *redis.Config
	Cache       *rediscache.Config
}

func NewConfig() (*Config, error) {
	c := &Config{
		Logger: &log.Config{},
		DB: &pq.Config{
			DSN: "postgres://postgres:secret@postgres:5432/test",
			Migrate: &migrations.Config{
				Directory: "/internal/db/migrations/pg",
				Action:    "up",
			},
		},
		PrivateHTTP: &echo.Config{
			Address: "0.0.0.0:9100",
		},
		RabbitMQ: &rabbitmq.Config{
			PoolConfig: &rabbitmqpool.Config{
				DSN:                    "amqp://guest:guest@rabbitmq:5672",
				MaxOpenChannelsPerConn: 5,
				MaxIdleChannelsPerConn: 5,
				ChannelMaxLifetime:     15 * time.Second,
				ChannelMaxIdleTime:     15 * time.Second,
				MaxOpenConns:           5,
				MaxIdleConns:           5,
				ConnMaxLifetime:        15 * time.Second,
				ConnMaxIdleTime:        15 * time.Second,
			},
		},

		Consumer: &rabbitmqconsum.Config{
			ExchangeName:  "test.events.dev",
			RoutingKey:    "TEST_EVENTS",
			RabbitQueue:   "test.queue.dev",
			RabbitConsume: "garage-test",
		},

		Publisher: &rabbitmqpub.Config{
			ExchangeName: "testout.events.dev",
			RoutingKey:   "TEST_EVENTS",
		},

		Redis: &redis.Config{
			DSN: "redis://redis:6379",
		},

		Cache: &rediscache.Config{
			KeyPrefix: "garage-test",
			ClearTime: 30 * time.Second,
		},
	}

	if err := envconfig.Init(c); err != nil {
		return nil, errors.Wrap(err, "init config")
	}

	return c, nil
}
