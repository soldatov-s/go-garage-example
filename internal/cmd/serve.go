package cmd

import (
	"context"

	"github.com/pkg/errors"
	apiv1 "github.com/soldatov-s/go-garage-example/api/v1"
	"github.com/soldatov-s/go-garage-example/internal/cfg"
	"github.com/soldatov-s/go-garage/app"
	"github.com/soldatov-s/go-garage/log"
	"github.com/soldatov-s/go-garage/providers/echo"
	"github.com/soldatov-s/go-garage/providers/pq"
	"github.com/soldatov-s/go-garage/providers/rabbitmq"
	"github.com/soldatov-s/go-garage/providers/redis"
	"golang.org/x/sync/errgroup"
)

func Run() error {
	ctx := context.Background()
	runner, ctx := errgroup.WithContext(ctx)

	config, err := cfg.NewConfig()
	if err != nil {
		return errors.Wrap(err, "new config")
	}

	logger, err := log.NewLogger(ctx, config.Logger)
	if err != nil {
		return errors.Wrap(err, "new logger")
	}
	ctx = logger.Zerolog().WithContext(ctx)

	meta := app.NewMeta(&app.MetaDeps{
		Name:        appName,
		Builded:     builded,
		Version:     version,
		Hash:        hash,
		Description: description,
	})

	manager := app.NewManager(&app.ManagerDeps{
		Meta:       meta,
		Logger:     logger,
		ErrorGroup: runner,
	})

	// Create connection to PostgreSQL
	pqEnity, err := pq.NewEnity(ctx, "garage_pq", config.DB)
	if err != nil {
		return errors.Wrap(err, "pq new enity")
	}

	if errAdd := manager.Add(ctx, pqEnity); errAdd != nil {
		return errors.Wrap(errAdd, "add enity to manager")
	}

	// Create connection to Redis
	redisEnity, err := redis.NewEnity(ctx, "garage_redis", config.Redis)
	if err != nil {
		return errors.Wrap(err, "redis new enity")
	}

	// Create cache in redis
	cacheEnity, err := redisEnity.AddCache(ctx, config.Cache)
	if err != nil {
		return errors.Wrap(err, "new cache")
	}

	if errAdd := manager.Add(ctx, redisEnity); errAdd != nil {
		return errors.Wrap(errAdd, "add enity to manager")
	}

	// Create connection to RabbitMQ
	rabbitmqEnity, err := rabbitmq.NewEnity(ctx, "garage_rabbitmq", config.RabbitMQ)
	if err != nil {
		return errors.Wrap(err, "rabbitmq new enity")
	}

	// Create consumer
	consumerEnity, err := rabbitmqEnity.AddConsumer(ctx, config.Consumer)
	if err != nil {
		return errors.Wrap(err, "new consumer")
	}

	// Create publisher
	publisherEnity, err := rabbitmqEnity.AddPublisher(ctx, config.Publisher)
	if err != nil {
		return errors.Wrap(err, "new publisher")
	}

	if errAdd := manager.Add(ctx, rabbitmqEnity); errAdd != nil {
		return errors.Wrap(errAdd, "add enity to manager")
	}

	middlewares := echo.DefaultMiddlewares()
	middlewares = append(middlewares, echo.HydrationZerolog(ctx))
	echoEnityPrivate, err := echo.NewEnity(ctx, "garage_echo", config.PrivateHTTP, middlewares...)
	if err != nil {
		return errors.Wrap(err, "echo new enity")
	}

	if errAdd := manager.Add(ctx, echoEnityPrivate); errAdd != nil {
		return errors.Wrap(errAdd, "add enity to manager")
	}

	manager.SetStatsHTTPEnityName(echoEnityPrivate.GetFullName())

	privateAPIV1, err := echoEnityPrivate.APIGroup(ctx, cfg.V1, meta.BuildInfo(), apiv1.GetSwagger)
	if err != nil {
		return errors.Wrap(err, "add api group")
	}

	repositoryDeps := &apiv1.RepositoryDeps{
		Conn: pqEnity,
	}
	repository, err := apiv1.NewRepository(repositoryDeps)
	if err != nil {
		return errors.Wrap(err, "new repository")
	}

	cacheDeps := apiv1.CacheDeps{
		Cache: cacheEnity,
	}

	cache := apiv1.NewCache(&cacheDeps)

	consumerDeps := &apiv1.ConsumerDeps{
		Repository: repository,
		Cache:      cache,
		Publisher:  publisherEnity,
	}

	consumer := apiv1.NewConsumer(consumerDeps)

	handlerDeps := &apiv1.HandlerDeps{
		Repository: repository,
		Cache:      cache,
	}
	handler := apiv1.NewHandler(handlerDeps)

	apiv1.RegisterHandlers(privateAPIV1, handler)

	if err := manager.Start(ctx); err != nil {
		return errors.Wrap(err, "start application")
	}

	if err := consumerEnity.Subscribe(ctx, runner, consumer); err != nil {
		return errors.Wrap(err, "subscribe consumer")
	}

	if err := manager.OSSignalWaiter(ctx); err != nil {
		return errors.Wrap(err, "os signal waiter")
	}

	if err := manager.Loop(ctx); err != nil {
		return errors.Wrap(err, "application loop")
	}

	return nil
}
