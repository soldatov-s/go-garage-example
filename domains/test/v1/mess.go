package test

import (
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
	"github.com/soldatov-s/go-garage/utils"
)

type Messenger interface {
	rabbitmq.SubscribeOptions
}

type Mess struct {
	repo  Repository
	cache Cacher
	msgs  *rabbitmq.Enity
	log   *zerolog.Logger
}

var _ Messenger = new(Mess)

func NewMess(log *zerolog.Logger, msgs *rabbitmq.Enity, repo Repository, cache Cacher) *Mess {
	return &Mess{
		repo:  repo,
		cache: cache,
		log:   log,
		msgs:  msgs,
	}
}

func (m *Mess) Consume(data []byte) error {
	var request Consume
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	// Check that code not exist in cache
	var streamState string
	if err := m.cache.Get(request.Code, &streamState); err == nil {
		m.log.Debug().Msgf("find code %s in cache", request.Code)
		return nil
	}

	test, err := m.repo.GetByCode(request.Code)
	if err != nil {
		return err
	}

	var testPublish Publish

	testPublish.Code = test.Code
	testPublish.SendAt = utils.UnixTimeToUTZ(request.SendAt)

	return m.msgs.SendMessage(testPublish)
}

func (m *Mess) Shutdown() {}
