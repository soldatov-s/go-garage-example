package test

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/utils"
	"github.com/soldatov-s/go-garage/x/sql"
)

const productionTestTable = "production.test"

type RepositoryGateway interface {
	CreateTest(ctx context.Context, data *Enity) (*Enity, error)
	GetByID(ctx context.Context, id int64) (data *Enity, err error)
	GetByCode(ctx context.Context, code string) (data *Enity, err error)
	HardDeleteByID(ctx context.Context, id int64) (err error)
	SoftDeleteByID(ctx context.Context, id int64) (err error)
}

type Repository struct {
	db *pq.Enity
	h  *sql.Helper
}

func NewRepo(db *pq.Enity) *Repository {
	return &Repository{
		h:  &sql.Helper{},
		db: db,
	}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Enity, error) {
	data := &Enity{}

	if err := r.h.SelectByID(r.db.Conn, productionTestTable, id, data); err != nil {
		return nil, errors.Wrap(err, "select by id")
	}

	return data, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*Enity, error) {
	logger := zerolog.Ctx(ctx)

	data := &Enity{}

	if err := r.db.Conn.GetContext(ctx, data, utils.JoinStrings(" ", "select * from", productionTestTable, "where code=$1"), code); err != nil {
		return nil, errors.Wrap(err, "get from db")
	}

	logger.Debug().Msgf("data %+v", data)

	return data, nil
}

func (r *Repository) HardDeleteByID(ctx context.Context, id int64) error {
	if err := r.h.HardDeleteByID(r.db.Conn, productionTestTable, id); err != nil {
		return errors.Wrap(err, "hard delete by id")
	}
	return nil
}

func (r *Repository) SoftDeleteByID(ctx context.Context, id int64) error {
	if err := r.h.SoftDeleteByID(r.db.Conn, productionTestTable, id); err != nil {
		return errors.Wrap(err, "soft delete by id")
	}
	return nil
}

func (r *Repository) CreateTest(ctx context.Context, data *Enity) (*Enity, error) {
	data.CreateTimestamp()

	result, err := r.h.InsertInto(r.db.Conn, productionTestTable, data)
	if err != nil {
		return nil, errors.Wrap(err, "insert into")
	}

	return result.(*Enity), nil
}
