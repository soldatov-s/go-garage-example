package test

import (
	"context"
	"encoding/json"

	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
	"github.com/soldatov-s/go-garage/utils"
)

type Messenger interface {
	rabbitmq.SubscribeOptions
}

type MessDeps struct {
	Repository RepositoryGateway
	Cache      Cacher
	Msgs       *rabbitmq.Enity
}

type Mess struct {
	repo  RepositoryGateway
	cache Cacher
	msgs  *rabbitmq.Enity
}

func NewMess(deps *MessDeps) *Mess {
	return &Mess{
		repo:  deps.Repository,
		cache: deps.Cache,
		msgs:  deps.Msgs,
	}
}

func (m *Mess) Consume(data []byte) error {
	var request Consume
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	ctx := context.Background()

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

	testPublish.Code = test.Code
	testPublish.SendAt = utils.UnixTimeToUTZ(request.SendAt)

	return m.msgs.SendMessage(testPublish)
}

func (m *Mess) Shutdown() {}
