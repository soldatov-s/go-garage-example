package test

import (
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
)

type AppInterface interface {
	PostToCacheHandler(ec echo.Context) error
	GetHandler(ec echo.Context) error
	PostHandler(ec echo.Context) error
	DeleteHandler(ec echo.Context) error
}

type App struct {
	repo  Repository
	cache Cacher
}

func NewApp(repo Repository, cache Cacher) *App {
	return &App{repo: repo, cache: cache}
}

func (a *App) PostToCacheHandler(ec echo.Context) error {
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
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	data, err := a.repo.GetByID(ID)
	if err != nil {
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	if err := a.cache.Set(data.Code, data); err != nil {
		log.Err(err).Msgf("internal server error, data %+v", data)
		return ec.InternalServerError(err)
	}

	if err := a.cache.Ping(); err != nil {
		log.Debug().Msgf("ping redis %s", err)
	}

	return ec.OkResult()
}

func (a *App) GetHandler(ec echo.Context) error {
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
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	data, err := a.repo.GetByID(ID)
	if err != nil {
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	return ec.OK(DataResult{Body: data})
}

func (a *App) PostHandler(ec echo.Context) error {
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
		log.Err(err).Msg("bad request")
		return ec.BadRequest(err)
	}

	data, err := a.repo.CreateTest(request)
	if err != nil {
		log.Err(err).Msgf("create data failed %+v", &request)
		return ec.CreateFailed(err)
	}

	return ec.OK(DataResult{Body: data})
}

func (a *App) DeleteHandler(ec echo.Context) error {
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
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
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
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	return ec.OkResult()
}
