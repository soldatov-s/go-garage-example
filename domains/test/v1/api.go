package test

import (
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
)

type APIInterface interface {
	SetRoutes(log *zerolog.Logger, publicHTTP, privateHTTP *echo.Enity, version string) error
	PostToCacheHandler(ec echo.Context) error
	GetHandler(ec echo.Context) error
	PostHandler(ec echo.Context) error
	DeleteHandler(ec echo.Context) error
}

type API struct {
	repo  Repository
	cache Cacher
}

var _ APIInterface = new(API)

func NewAPI(repo Repository, cache Cacher) *API {
	return &API{repo: repo, cache: cache}
}

func (a *API) SetRoutes(log *zerolog.Logger, publicHTTP, privateHTTP *echo.Enity, version string) error {
	publicV1, err := publicHTTP.GetAPIVersionGroup(version)
	if err != nil {
		return errors.Wrap(err, "get version group")
	}

	publicV1.Group.Use(echo.HydrationLogger(log))
	publicV1.Group.POST("/test/:id", echo.Handler(a.PostToCacheHandler))

	privateV1, err := privateHTTP.GetAPIVersionGroup(version)
	if err != nil {
		return errors.Wrap(err, "get version group")
	}

	privateV1.Group.Use(echo.HydrationLogger(log))
	privateV1.Group.GET("/test/:id", echo.Handler(a.GetHandler))
	privateV1.Group.POST("/test", echo.Handler(a.PostHandler))
	privateV1.Group.DELETE("/test/:id", echo.Handler(a.DeleteHandler))

	return nil
}

func (a *API) PostToCacheHandler(ec echo.Context) error {
	// Swagger
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

	log := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		log.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	data, err := a.repo.GetByID(ID)
	if err != nil {
		log.Err(err).Msgf("failed to get data by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to get data by id"))
	}

	if err := a.cache.Set(data.Code, data); err != nil {
		log.Err(err).Msgf("failed to save to cache, data %+v", data)
		return ec.InternalServerError(errors.Wrap(err, "failed to save to cache"))
	}

	if err := a.cache.Ping(); err != nil {
		log.Debug().Msgf("ping redis %s", err)
	}

	if err := ec.OkResult(); err != nil {
		log.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}

func (a *API) GetHandler(ec echo.Context) error {
	// Swagger
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

	log := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		log.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	data, err := a.repo.GetByID(ID)
	if err != nil {
		log.Err(err).Msgf("failed to get data by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to get data"))
	}

	if err := ec.OK(DataResult{Body: data}); err != nil {
		log.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}

func (a *API) PostHandler(ec echo.Context) error {
	// Swagger
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

	// Main code of handler
	log := ec.GetLog()

	request := &Enity{}

	if err := ec.Bind(request); err != nil {
		log.Err(err).Msg("failed to bind body")
		return ec.BadRequest(errors.Wrap(err, "failed to bind body"))
	}

	data, err := a.repo.CreateTest(request)
	if err != nil {
		log.Err(err).Msgf("failed create data %+v", &request)
		return ec.CreateFailed(errors.Wrap(err, "failed create data"))
	}

	return ec.OK(DataResult{Body: data})
}

func (a *API) DeleteHandler(ec echo.Context) error {
	// Swagger
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

	log := ec.GetLog()

	ID, err := ec.GetInt64Param("id")
	if err != nil {
		log.Err(err).Msgf("failed to convert id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to convert id"))
	}

	hard := ec.QueryParam("hard")
	switch hard {
	case "true":
		err = a.repo.HardDeleteByID(ID)
	case "false":
		err = a.repo.SoftDeleteByID(ID)
	default:
		err = errors.New("unknown value for hard parameter")
	}

	if err != nil {
		log.Err(err).Msgf("failed to delete by id %q", ec.Param("id"))
		return ec.BadRequest(errors.Wrap(err, "failed to delete"))
	}

	if err := ec.OkResult(); err != nil {
		log.Err(err).Msg("failed to write answer")
		return ec.InternalServerError(errors.Wrap(err, "failed to write answer"))
	}

	return nil
}
