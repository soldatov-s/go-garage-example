package testv1

import (
	"github.com/soldatov-s/go-garage-example/models"
	"github.com/soldatov-s/go-garage/providers/db"
	"github.com/soldatov-s/go-garage/x/sql"
)

func (t *TestV1) GetTestByID(id int64) (data *models.Test, err error) {
	data = &models.Test{}

	if err := sql.SelectByID(t.db.Conn, "production.test", id, data); err != nil {
		return nil, err
	}

	t.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (t *TestV1) GetTestByCode(code string) (data *models.Test, err error) {
	if t.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	data = &models.Test{}

	err = t.db.Conn.Get(data, "select * from production.test where code=$1", code)
	if err != nil {
		return nil, err
	}

	t.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (t *TestV1) HardDeleteTestByID(id int64) (err error) {
	return sql.HardDeleteByID(t.db.Conn, "production.test", id)
}

func (t *TestV1) SoftDeleteTestByID(id int64) (err error) {
	return sql.SoftDeleteByID(t.db.Conn, "production.test", id)
}

func (t *TestV1) CreateTest(data *models.Test) (*models.Test, error) {
	data.CreateTimestamp()

	result, err := sql.InsertInto(t.db.Conn, "production.test", data)
	if err != nil {
		return nil, err
	}

	return result.(*models.Test), nil
}
