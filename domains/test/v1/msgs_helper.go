package testv1

import (
	"encoding/json"

	"github.com/soldatov-s/go-garage-example/models"
	"github.com/soldatov-s/go-garage/utils"
)

func (t *TestV1) ConsumeHndl(data []byte) error {
	var request models.TestConsume
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	// Check that code not exist in cache
	var streamState string
	if err := t.cache.Get(request.Code, &streamState); err == nil {
		t.log.Debug().Msgf("find code %s in cache", request.Code)
		return nil
	}

	test, err := t.GetTestByCode(request.Code)
	if err != nil {
		return err
	}

	var testPublish models.TestPublish

	testPublish.Code = test.Code
	testPublish.SendAt = utils.UnixTimeToUTZ(request.SendAt)

	return t.msgs.SendMessage(testPublish)
}
