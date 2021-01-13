package testv1

import (
	// stdlib
	"errors"
	"strings"

	// local
	"github.com/soldatov-s/go-garage-example/models"

	// other
	"github.com/soldatov-s/go-garage/utils"
)

var (
	ErrDBConnNotEstablished = errors.New("database connection not established")
)

func (t *TestV1) GetTestByID(id int) (data *models.Test, err error) {
	data = &models.Test{}

	if t.db.Conn == nil {
		return nil, ErrDBConnNotEstablished
	}

	err = t.db.Conn.Get(data, "select * from production.test where id=$1", id)
	if err != nil {
		return nil, err
	}

	t.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (t *TestV1) GetTestByCode(code string) (data *models.Test, err error) {
	data = &models.Test{}

	if t.db.Conn == nil {
		return nil, ErrDBConnNotEstablished
	}

	err = t.db.Conn.Get(data, "select * from production.test where code=$1", code)
	if err != nil {
		return nil, err
	}

	t.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (t *TestV1) HardDeleteTestByID(id int) (err error) {
	if t.db.Conn == nil {
		return ErrDBConnNotEstablished
	}

	_, err = t.db.Conn.Exec(t.db.Conn.Rebind("delete from production.test where id=$1"), id)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestV1) SoftDeleteTestByID(id int) (err error) {
	data, err := t.GetTestByID(id)
	if err != nil {
		return err
	}

	if data.DeletedAt.Valid {
		return err
	}

	data.DeletedAt.Timestamp()

	query := make([]string, 0, len(data.SQLParamsRequest()))
	for _, param := range data.SQLParamsRequest() {
		query = append(query, param+"=:"+param)
	}

	if t.db.Conn == nil {
		return ErrDBConnNotEstablished
	}

	_, err = t.db.Conn.NamedExec(
		t.db.Conn.Rebind(utils.JoinStrings(" ", "UPDATE production.test SET", strings.Join(query, ", "), "WHERE id=:id")),
		data)

	if err != nil {
		return err
	}

	return nil
}

func (t *TestV1) CreateTest(data *models.Test) (*models.Test, error) {
	data.CreatedAt.Timestamp()
	data.UpdatedAt = data.CreatedAt

	if t.db.Conn == nil {
		return nil, ErrDBConnNotEstablished
	}

	stmt, err := t.db.Conn.PrepareNamed(
		t.db.Conn.Rebind(utils.JoinStrings(" ", "INSERT INTO production.test", "("+strings.Join(data.SQLParamsRequest(), ", ")+")",
			"VALUES", "("+":"+strings.Join(data.SQLParamsRequest(), ", :")+") returning *")))
	if err != nil {
		return nil, err
	}

	err = stmt.Get(data, data)
	stmt.Close()

	return data, err
}
