package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/db"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/utils"
	"github.com/soldatov-s/go-garage/x/sql"
)

type Repository interface {
	CreateTest(data *Enity) (*Enity, error)
	GetByID(id int64) (data *Enity, err error)
	GetByCode(code string) (data *Enity, err error)
	HardDeleteByID(id int64) (err error)
	SoftDeleteByID(id int64) (err error)
}

const productionTestTable = "production.test"

type RepoConfig struct {
	DBName string
}

type Repo struct {
	db  *pq.Enity
	ctx context.Context
	log zerolog.Logger
}

func NewRepository(ctx context.Context, cfg *RepoConfig) (*Repo, error) {
	r := &Repo{}
	var err error
	if r.db, err = pq.GetEnityTypeCast(ctx, cfg.DBName); err != nil {
		return nil, errors.Wrap(err, "failed to get pq enity")
	}

	r.log = logger.GetPackageLogger(ctx, empty{})

	return r, nil
}

func (r *Repo) GetByID(id int64) (*Enity, error) {
	data := &Enity{}

	if err := sql.SelectByID(r.db.Conn, productionTestTable, id, data); err != nil {
		return nil, errors.Wrap(err, "select by id")
	}

	r.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (r *Repo) GetByCode(code string) (*Enity, error) {
	if r.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	data := &Enity{}

	if err := r.db.Conn.Get(data, utils.JoinStrings(" ", "select * from", productionTestTable, "where code=$1"), code); err != nil {
		return nil, errors.Wrap(err, "get from db")
	}

	r.log.Debug().Msgf("data %+v", data)

	return data, nil
}

func (r *Repo) HardDeleteByID(id int64) error {
	if err := sql.HardDeleteByID(r.db.Conn, productionTestTable, id); err != nil {
		return errors.Wrap(err, "hard delete by id")
	}
	return nil
}

func (r *Repo) SoftDeleteByID(id int64) error {
	if err := sql.SoftDeleteByID(r.db.Conn, productionTestTable, id); err != nil {
		return errors.Wrap(err, "soft delete by id")
	}
	return nil
}

func (r *Repo) CreateTest(data *Enity) (*Enity, error) {
	data.CreateTimestamp()

	if r.ctx == nil {
		r.ctx, _ = sql.Create(context.Background())
	}

	result, err := sql.InsertInto(r.ctx, r.db.Conn, productionTestTable, data)
	if err != nil {
		return nil, errors.Wrap(err, "insert into")
	}

	return result.(*Enity), nil
}
