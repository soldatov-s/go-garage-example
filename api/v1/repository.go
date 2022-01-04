package apiv1

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/pq"
	"github.com/soldatov-s/go-garage/x/sqlx/miniorm"
	"github.com/soldatov-s/go-garage/x/stringsx"
)

const productionTestTable = "production.test"

type RepositoryDeps struct {
	Conn *pq.Enity
}

type Repository struct {
	conn *pq.Enity
	orm  *miniorm.ORM
}

func NewRepository(deps *RepositoryDeps) (*Repository, error) {
	r := &Repository{
		conn: deps.Conn,
	}

	ormDeps := &miniorm.ORMDeps{
		Conn:             deps.Conn,
		Target:           productionTestTable,
		Data:             &Enity{},
		PredefinedFields: miniorm.NewPredefinedFields(),
	}
	var err error
	r.orm, err = miniorm.NewORM(ormDeps)
	if err != nil {
		return nil, errors.Wrap(err, "new miniorm")
	}

	return r, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Enity, error) {
	data := &Enity{}

	if err := r.orm.SelectByID(ctx, id, data); err != nil {
		return nil, errors.Wrap(err, "select by id")
	}

	return data, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*Enity, error) {
	logger := zerolog.Ctx(ctx)

	data := &Enity{}

	if err := r.conn.GetConn().GetContext(ctx, data,
		stringsx.JoinStrings(" ", "select * from", productionTestTable, "where code=$1"), code); err != nil {
		return nil, errors.Wrap(err, "get from db")
	}

	logger.Debug().Msgf("data %+v", data)

	return data, nil
}

func (r *Repository) HardDeleteByID(ctx context.Context, id int64) error {
	if err := r.orm.HardDeleteByID(ctx, id); err != nil {
		return errors.Wrap(err, "hard delete by id")
	}
	return nil
}

func (r *Repository) SoftDeleteByID(ctx context.Context, id int64) error {
	if err := r.orm.SoftDeleteByID(ctx, id); err != nil {
		return errors.Wrap(err, "soft delete by id")
	}
	return nil
}

func (r *Repository) CreateTest(ctx context.Context, data *Enity) (*Enity, error) {
	data.CreatedAt = NewNullTime()
	data.UpdatedAt = data.CreatedAt

	if err := r.orm.InsertInto(ctx, data, data); err != nil {
		return nil, errors.Wrap(err, "insert into")
	}

	return data, nil
}
