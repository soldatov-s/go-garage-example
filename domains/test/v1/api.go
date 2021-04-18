package test

import (
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
)

type HandlerGateway interface {
	SetRoutes(publicGroup, privateGroup *echo.Group)
	PostToCacheHandler(ec echo.Context) error
	GetHandler(ec echo.Context) error
	PostHandler(ec echo.Context) error
	DeleteHandler(ec echo.Context) error
}

type HandlerDeps struct {
	Repository RepositoryGateway
	Cache      Cacher
}

type Handler struct {
	repository RepositoryGateway
	cache      Cacher
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{repository: deps.Repository, cache: deps.Cache}
}

func (a *Handler) SetRoutes(publicGroup, privateGroup *echo.Group) {
	publicGroup.POST("/test/:id", echo.Handler(a.PostToCacheHandler))

	privateGroup.GET("/test/:id", echo.Handler(a.GetHandler))
	privateGroup.POST("/test", echo.Handler(a.PostHandler))
	privateGroup.DELETE("/test/:id", echo.Handler(a.DeleteHandler))
}

func (a *Handler) PostToCacheHandler(ec echo.Context) error {
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler put data to cache for requested ID").
			SetSummary("Put data to cache by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusInternalServerError, "INTERNAL SERVER ERROR", httpsrv.ErrorAnsw{})

		return nil
	}

	ctx := ec.Request().Context()
	logger := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		logger.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	data, err := a.repository.GetByID(ctx, ID)
	if err != nil {
		logger.Err(err).Msgf("failed to get data by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to get data by id"))
	}

	if err := a.cache.Set(ctx, data.Code, data); err != nil {
		logger.Err(err).Msgf("failed to save to cache, data %+v", data)
		return ec.InternalServerError(errors.Wrap(err, "failed to save to cache"))
	}

	if err := a.cache.Ping(ctx); err != nil {
		logger.Debug().Msgf("ping redis %s", err)
	}

	if err := ec.OkResult(); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}

func (a *Handler) GetHandler(ec echo.Context) error {
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler getting data for requested ID").
			SetSummary("Get data by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddResponse(http.StatusOK, "Data", DataResult{Body: Enity{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.ErrorAnsw{})

		return nil
	}

	ctx := ec.Request().Context()
	logger := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		logger.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	data, err := a.repository.GetByID(ctx, ID)
	if err != nil {
		logger.Err(err).Msgf("failed to get data by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to get data"))
	}

	if err := ec.OK(DataResult{Body: data}); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}

func (a *Handler) PostHandler(ec echo.Context) error {
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler create new data").
			SetSummary("Create Data Handler").
			AddInBodyParameter("data", "Data", Enity{}, true).
			AddResponse(http.StatusOK, "Data", &DataResult{Body: Enity{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusConflict, "CREATE DATA FAILED", httpsrv.ErrorAnsw{})

		return nil
	}

	ctx := ec.Request().Context()
	logger := ec.GetLog()

	request := &Enity{}

	if err := ec.Bind(request); err != nil {
		logger.Err(err).Msg("failed to bind body")
		return ec.BadRequest(errors.Wrap(err, "failed to bind body"))
	}

	data, err := a.repository.CreateTest(ctx, request)
	if err != nil {
		logger.Err(err).Msgf("failed create data %+v", &request)
		return ec.CreateFailed(errors.Wrap(err, "failed create data"))
	}

	return ec.OK(DataResult{Body: data})
}

func (a *Handler) DeleteHandler(ec echo.Context) error {
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler deletes data for requested ID").
			SetSummary("Delete data by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddInQueryParameter("hard", "Hard delete data, if equal true, delete hard", reflect.Bool, false).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.ErrorAnsw{})

		return nil
	}

	ctx := ec.Request().Context()
	logger := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		logger.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	hard := ec.QueryParam("hard")
	switch hard {
	case "true":
		err = a.repository.HardDeleteByID(ctx, ID)
	case "false":
		err = a.repository.SoftDeleteByID(ctx, ID)
	default:
		err = errors.New("unknown value for hard parameter")
	}

	if err != nil {
		logger.Err(err).Msgf("failed to delete by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to delete"))
	}

	if err := ec.OkResult(); err != nil {
		logger.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}
