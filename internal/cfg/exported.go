package cfg

import (
	"context"
	"time"

	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/config"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
	"github.com/soldatov-s/go-garage/providers/stats/garage"
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
	Logger      *logger.Config
	DB          *pq.Config
	PublicHTTP  *echo.Config
	PrivateHTTP *echo.Config
	RabbitMQ    *rabbitmq.Config
	Redis       *redis.Config
	Stats       *garage.Config
}

func Get(ctx context.Context) *Config {
	return config.Get(ctx).Service.(*Config)
}

func NewConfig() *Config {
	return &Config{
		Logger: &logger.Config{},
		DB: &pq.Config{
			DSN: "postgres://postgres:secret@postgres:5432/test",
			Migrate: &pq.MigrateConfig{
				Directory: "/internal/db/migrations/pg",
				Action:    "up",
			},
		},
		PublicHTTP: &echo.Config{
			Address: "0.0.0.0:9000",
		},
		PrivateHTTP: &echo.Config{
			Address: "0.0.0.0:9100",
		},
		RabbitMQ: &rabbitmq.Config{
			DSN: "amqp://guest:guest@rabbitmq:5672",
			Consumer: &rabbitmq.ConsumerConfig{
				RabbitBaseConfig: rabbitmq.RabbitBaseConfig{
					ExchangeName: "test.events.dev",
					RoutingKey:   "TEST_EVENTS",
				},
				RabbitQueue:   "test.queue.dev",
				RabbitConsume: "garage-test",
			},
			Publisher: &rabbitmq.PublisherConfig{
				RabbitBaseConfig: rabbitmq.RabbitBaseConfig{
					ExchangeName: "testout.events.dev",
					RoutingKey:   "TEST_EVENTS",
				},
			},
		},
		Redis: &redis.Config{
			DSN:       "redis://redis:6379",
			KeyPrefix: "garage-test_",
			ClearTime: 30 * time.Second,
		},
		Stats: &garage.Config{
			HTTPProviderName: echo.DefaultProviderName,
			HTTPEnityName:    "private",
		},
	}
}
