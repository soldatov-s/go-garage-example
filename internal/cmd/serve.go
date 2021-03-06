package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	testv1 "github.com/soldatov-s/go-garage-example/domains/test/v1"
	"github.com/soldatov-s/go-garage-example/internal/cfg"
	"github.com/soldatov-s/go-garage/app"
	"github.com/soldatov-s/go-garage/meta"
	"github.com/soldatov-s/go-garage/providers/cache/redis"
	"github.com/soldatov-s/go-garage/providers/config/envconfig"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
	"github.com/soldatov-s/go-garage/providers/stats"
	"github.com/soldatov-s/go-garage/providers/stats/garage"
	"github.com/spf13/cobra"
)

// Private type, used for configure logger
type Empty struct{}

func addMetrics(ctx context.Context) error {
	s, err := garage.GetEnityTypeCast(ctx, cfg.StatsName)
	if err != nil {
		return errors.Wrap(err, "get stat enity")
	}

	if err := s.RegisterReadyCheck("TEST",
		func() (bool, string) {
			return true, "test error"
		}); err != nil {
		return errors.Wrap(err, "registrate ready check")
	}

	metricOptions := &stats.MetricOptions{
		Metric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "test_alive",
				Help: "test server link",
			}),
		MetricFunc: func(m interface{}) {
			(m.(prometheus.Gauge)).Set(1)
		},
	}

	if err := s.RegisterMetric(
		"test",
		metricOptions); err != nil {
		return errors.Wrap(err, "registarte metric")
	}

	return nil
}

func initService() context.Context {
	// Create context
	ctx := app.CreateAppContext(context.Background())

	// Set app info
	ctx = meta.SetAppInfo(ctx, appName, builded, hash, version, description)

	// Load and parse config
	ctx, err := envconfig.RegistrateAndParse(ctx, cfg.NewConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parsing config")
	}
	log.Info().Msg("configuration parsed successfully")
	c := cfg.Get(ctx)

	// Registrate logger
	ctx = logger.RegistrateAndInitilize(ctx, c.Logger)

	// Get logger for package
	packageLogger := logger.GetPackageLogger(ctx, Empty{})

	a := meta.Get(ctx)
	packageLogger.Info().Msgf("starting %s (%s)...", a.Name, a.GetBuildInfo())
	packageLogger.Info().Msg(a.Description)

	return ctx
}

func initProviders(ctx context.Context) context.Context {
	var err error
	c := cfg.Get(ctx)
	packageLogger := logger.GetPackageLogger(ctx, Empty{})

	// Initialize providers
	ctx, err = pq.RegistrateEnity(ctx, cfg.DBName, c.DB)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed create db connection")
	}

	ctx, err = redis.RegistrateEnity(ctx, cfg.CacheName, c.Redis)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed create cache connection")
	}

	ctx, err = rabbitmq.RegistrateEnity(ctx, cfg.MsgsName, c.RabbitMQ)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to create msgs connection")
	}

	// Public HTTP
	ctx, err = echo.RegistrateEnity(ctx, cfg.PublicHTTP, c.PublicHTTP)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to registrate http")
	}

	publicEchoEnity, err := echo.GetEnityTypeCast(ctx, cfg.PublicHTTP)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to get http")
	}

	publicEchoEnity.Server.Use(echo.CORSDefault(), echo.HydrationRequestID())

	if err = publicEchoEnity.CreateAPIVersionGroup(cfg.V1); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to create api group")
	}

	// Private HTTP
	ctx, err = echo.RegistrateEnity(ctx, cfg.PrivateHTTP, c.PrivateHTTP)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to registrate http")
	}

	privateEchoEnity, err := echo.GetEnityTypeCast(ctx, cfg.PrivateHTTP)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to get http")
	}

	privateEchoEnity.Server.Use(echo.CORSDefault(), echo.HydrationRequestID())

	if err = privateEchoEnity.CreateAPIVersionGroup(cfg.V1); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to create api group")
	}

	if ctx, err = garage.RegistrateEnity(ctx, cfg.StatsName, c.Stats); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to registrate stats")
	}

	if err = addMetrics(ctx); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to registrate metrics")
	}

	return ctx
}

func initDomains(ctx context.Context) context.Context {
	var err error
	packageLogger := logger.GetPackageLogger(ctx, Empty{})

	// Initilize domains
	if ctx, err = testv1.Registrate(ctx, &testv1.Config{
		DBName:      cfg.DBName,
		CacheName:   cfg.CacheName,
		MsgsName:    cfg.MsgsName,
		PublicHTTP:  cfg.PublicHTTP,
		PrivateHTTP: cfg.PrivateHTTP,
		Version:     "1",
	}); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to create domain testv1")
	}

	// Subscribe domain to rabbitmq
	rabbimqEnity, err := rabbitmq.GetEnityTypeCast(ctx, cfg.MsgsName)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to get msgs enity")
	}

	testV1, err := testv1.Get(ctx)
	if err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to get testv1 domain")
	}

	if err = rabbimqEnity.Subscribe(testV1.GetMess()); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to subscribe messaging")
	}

	return ctx
}

func serveHandler(_ *cobra.Command, _ []string) {
	ctx := initService()
	packageLogger := logger.GetPackageLogger(ctx, Empty{})

	ctx = initProviders(ctx)
	ctx = initDomains(ctx)

	// Start connect
	if err := app.Start(ctx); err != nil {
		packageLogger.Fatal().Err(err).Msg("failed to start service")
	}

	app.Loop(ctx)
}
