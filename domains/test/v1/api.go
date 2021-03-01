package testv1

import (
	"net/http"
	"reflect"

	"github.com/soldatov-s/go-garage-example/models"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
)

func (t *TestV1) testPostToCacheHandler(ec echo.Context) (err error) {
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

	data, err := t.GetTestByID(ID)
	if err != nil {
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	if err := t.cache.Set(data.Code, data); err != nil {
		log.Err(err).Msgf("internal server error, data %+v", data)
		return ec.InternalServerError(err)
	}

	if _, err := t.cache.Conn.Ping(t.ctx).Result(); err != nil {
		log.Debug().Msgf("ping redis %s", err)
	}

	return ec.OkResult()
}

func (t *TestV1) testGetHandler(ec echo.Context) (err error) {
	// Swagger
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler getting data for requested ID").
			SetSummary("Get data by ID").
			AddInPathParameter("id", "ID", reflect.Int).
			AddResponse(http.StatusOK, "Data", TestDataResult{Body: models.Test{}}).
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

	data, err := t.GetTestByID(ID)
	if err != nil {
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	return ec.OK(TestDataResult{Body: data})
}

func (t *TestV1) testPostHandler(ec echo.Context) (err error) {
	// Swagger
	if ec.IsBuildingSwagger() {
		ec.AddToSwagger().
			SetProduces("application/json").
			SetDescription("This handler create new data").
			SetSummary("Create Data Handler").
			AddInBodyParameter("data", "Data", models.Test{}, true).
			AddResponse(http.StatusOK, "Data", &TestDataResult{Body: models.Test{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.ErrorAnsw{}).
			AddResponse(http.StatusConflict, "CREATE DATA FAILED", httpsrv.ErrorAnsw{})

		return nil
	}

	// Main code of handler
	log := ec.GetLog()

	var request models.Test

	err = ec.Bind(&request)
	if err != nil {
		log.Err(err).Msg("bad request")
		return ec.BadRequest(err)
	}

	data, err := t.CreateTest(&request)
	if err != nil {
		log.Err(err).Msgf("create data failed %+v", &request)
		return ec.CreateFailed(err)
	}

	return ec.OK(TestDataResult{Body: data})
}

func (t *TestV1) testDeleteHandler(ec echo.Context) (err error) {
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
	if hard == "true" {
		err = t.HardDeleteTestByID(ID)
	} else {
		err = t.SoftDeleteTestByID(ID)
	}

	if err != nil {
		log.Err(err).Msgf("bad request, id %s", ec.Param("id"))
		return ec.BadRequest(err)
	}

	return ec.OkResult()
}
