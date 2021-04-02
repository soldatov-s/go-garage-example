package test

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/msgs/rabbitmq"
	"github.com/soldatov-s/go-garage/utils"
)

type Messenger interface {
	ConsumeHndl(data []byte) error
}

type Mess struct {
	repo  Repository
	cache Cacher
	msgs  *rabbitmq.Enity
	log   zerolog.Logger
}

func NewMess(ctx context.Context, msgsName string, repo Repository, cache Cacher) (*Mess, error) {
	m := &Mess{repo: repo, cache: cache}

	var err error
	if m.msgs, err = rabbitmq.GetEnityTypeCast(ctx, msgsName); err != nil {
		return nil, errors.Wrap(err, "failed to get rabbitmq enity")
	}

	m.log = logger.GetPackageLogger(ctx, empty{})

	return m, nil
}

func (m *Mess) ConsumeHndl(data []byte) error {
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
