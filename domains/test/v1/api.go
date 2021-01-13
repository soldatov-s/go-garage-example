package testv1

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/soldatov-s/go-garage-example/models"

	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	echoSwagger "github.com/soldatov-s/go-swagger/echo-swagger"
)

func (t *TestV1) testPostToCacheHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = errors.New("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("This handler put data to cache for requested ID").
			SetSummary("Put data to cache by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err)).
			AddResponse(http.StatusInternalServerError, "INTERNAL SERVER ERROR", httpsrv.InternalServerError(err))

		return nil
	}

	log := echo.GetLog(ec)

	ID, err := strconv.Atoi(ec.Param("id"))
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	data, err := t.GetTestByID(ID)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	if err := t.cache.Set(data.Code, data); err != nil {
		log.Err(err).Msgf("INTERNAL SERVER ERROR, data %+v", data)
		return echo.InternalServerError(ec, err)
	}

	if _, err := t.cache.Conn.Ping().Result(); err != nil {
		log.Debug().Msgf("ping redis %s", err)
	}

	return echo.OkResult(ec)
}

func (t *TestV1) testGetHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = errors.New("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("This handler getting data for requested ID").
			SetSummary("Get data by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddResponse(http.StatusOK, "Data", TestDataResult{Body: models.Test{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err))

		return nil
	}

	log := echo.GetLog(ec)

	ID, err := strconv.Atoi(ec.Param("id"))
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	data, err := t.GetTestByID(ID)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	return echo.OK(ec, TestDataResult{Body: data})
}

func (t *TestV1) testPostHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = errors.New("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("This handler create new data").
			SetSummary("Create Data Handler").
			AddInBodyParameter("data", "Data", models.Test{}, true).
			AddResponse(http.StatusOK, "Data", &TestDataResult{Body: models.Test{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "CREATE DATA FAILED", httpsrv.CreateFailed(err))

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	var request models.Test

	err = ec.Bind(&request)
	if err != nil {
		log.Err(err).Msg("BAD REQUEST")
		return echo.BadRequest(ec, err)
	}

	data, err := t.CreateTest(&request)
	if err != nil {
		log.Err(err).Msgf("CREATE DATA FAILED %+v", &request)
		return echo.CreateFailed(ec, err)
	}

	return echo.OK(ec, TestDataResult{Body: data})
}

func (t *TestV1) testDeleteHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = errors.New("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("This handler deletes data for requested ID").
			SetSummary("Delete data by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddInQueryParameter("hard", "Hard delete user, if equal true, delete hard", reflect.Bool, false).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err))

		return nil
	}

	log := echo.GetLog(ec)

	ID, err := strconv.Atoi(ec.Param("id"))
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	hard := ec.QueryParam("hard")
	if hard == "true" {
		err = t.HardDeleteTestByID(ID)
	} else {
		err = t.SoftDeleteTestByID(ID)
	}

	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))
		return echo.BadRequest(ec, err)
	}

	return echo.OkResult(ec)
}
