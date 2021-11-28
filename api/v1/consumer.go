package apiv1

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	rabbitmqpub "github.com/soldatov-s/go-garage/providers/rabbitmq/publisher"
	"github.com/soldatov-s/go-garage/x/timex"
)

type ConsumerRepositoryGateway interface {
	GetByCode(ctx context.Context, code string) (*Enity, error)
}

type ConsumerCacherGateway interface {
	Get(ctx context.Context, key string, value *string) error
	Set(ctx context.Context, key string, value *Enity) error
}

type ConsumerDeps struct {
	Repository ConsumerRepositoryGateway
	Cache      ConsumerCacherGateway
	Publisher  *rabbitmqpub.Publisher
}

type Consumer struct {
	repo      ConsumerRepositoryGateway
	cache     ConsumerCacherGateway
	publisher *rabbitmqpub.Publisher
}

func NewConsumer(deps *ConsumerDeps) *Consumer {
	return &Consumer{
		repo:      deps.Repository,
		cache:     deps.Cache,
		publisher: deps.Publisher,
	}
}

func (m *Consumer) Consume(ctx context.Context, data []byte) error {
	var request Consume
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	// Check that code not exist in cache
	var streamState string
	if err := m.cache.Get(ctx, request.Code, &streamState); err == nil {
		return nil
	}

	test, err := m.repo.GetByCode(ctx, request.Code)
	if err != nil {
		return err
	}

	var testPublish Publish

	testPublish.Code = *test.Code
	testPublish.SendAt = timex.UnixTimeToUTZ(request.SendAt)

	if err := m.publisher.SendMessage(ctx, testPublish); err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

func (m *Consumer) Shutdown(ctx context.Context) error {
	return nil
}
