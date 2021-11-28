package apiv1

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	garageEcho "github.com/soldatov-s/go-garage/providers/echo"
	"github.com/soldatov-s/go-garage/x/httpx"
)

type HandlerCacheGateway interface {
	Set(ctx context.Context, key string, value *Enity) error
}

type RepositoryGateway interface {
	CreateTest(ctx context.Context, data *Enity) (*Enity, error)
	GetByID(ctx context.Context, id int64) (data *Enity, err error)
	GetByCode(ctx context.Context, code string) (data *Enity, err error)
	HardDeleteByID(ctx context.Context, id int64) (err error)
	SoftDeleteByID(ctx context.Context, id int64) (err error)
}

type HandlerDeps struct {
	Repository RepositoryGateway
	Cache      HandlerCacheGateway
}

type Handler struct {
	repository RepositoryGateway
	cache      HandlerCacheGateway
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		repository: deps.Repository,
		cache:      deps.Cache,
	}
}

func (a *Handler) PostToCacheHandler(ec echo.Context, id int64) error {
	ctx := ec.Request().Context()
	logger := garageEcho.GetZerologger(ec)

	data, err := a.repository.GetByID(ctx, id)
	if err != nil {
		logger.Err(err).Msgf("failed to get data by id %q", id)
		return ec.JSON(http.StatusBadRequest, httpx.BadRequest(errors.Wrap(err, "failed to get data by id")))
	}

	if err := a.cache.Set(ctx, *data.Code, data); err != nil {
		logger.Err(err).Msgf("failed to save to cache, data %+v", data)
		return ec.JSON(http.StatusInternalServerError, httpx.InternalServerError(errors.Wrap(err, "failed to save to cache")))
	}

	if err := ec.JSON(http.StatusOK, httpx.OkResult()); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.JSON(http.StatusInternalServerError, httpx.InternalServerError(errors.Wrap(err, "failed to write answer")))
	}

	return nil
}

func (a *Handler) GetHandler(ec echo.Context, id int64) error {
	ctx := ec.Request().Context()
	logger := garageEcho.GetZerologger(ec)

	data, err := a.repository.GetByID(ctx, id)
	if err != nil {
		logger.Err(err).Msgf("failed to get data by id %q", id)
		return ec.JSON(http.StatusBadRequest, httpx.BadRequest(errors.Wrap(err, "failed to get data")))
	}

	if err := ec.JSON(http.StatusOK, DataResult{Result: data}); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.JSON(http.StatusInternalServerError, httpx.InternalServerError(errors.Wrap(err, "failed to write answer")))
	}

	return nil
}

func (a *Handler) PostHandler(ec echo.Context) error {
	ctx := ec.Request().Context()
	logger := garageEcho.GetZerologger(ec)

	request := &Enity{}

	if errBind := ec.Bind(request); errBind != nil {
		logger.Err(errBind).Msg("failed to bind body")
		return ec.JSON(http.StatusBadRequest, httpx.BadRequest(errors.Wrap(errBind, "failed to bind body")))
	}

	data, err := a.repository.CreateTest(ctx, request)
	if err != nil {
		logger.Err(err).Msgf("failed create data %+v", &request)
		return ec.JSON(http.StatusConflict, httpx.Conflict(errors.Wrap(err, "failed create data")))
	}

	return ec.JSON(http.StatusOK, DataResult{Result: data})
}

func (a *Handler) DeleteHandler(ec echo.Context, id int64, params DeleteHandlerParams) error {
	ctx := ec.Request().Context()
	logger := garageEcho.GetZerologger(ec)

	var err error
	if *params.Hard {
		err = a.repository.HardDeleteByID(ctx, id)
	} else {
		err = a.repository.SoftDeleteByID(ctx, id)
	}

	if err != nil {
		logger.Err(err).Msgf("failed to delete by id %q", id)
		return ec.JSON(http.StatusBadRequest, httpx.BadRequest(errors.Wrap(err, "failed to delete")))
	}

	if err := ec.JSON(http.StatusOK, httpx.OkResult()); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.JSON(http.StatusInternalServerError, httpx.InternalServerError(errors.Wrap(err, "failed to write answer")))
	}

	return nil
}
