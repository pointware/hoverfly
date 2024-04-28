package v2

import (
	"fmt"
	"net/http"

	"io/ioutil"

	"github.com/SpectoLabs/hoverfly/core/handlers"
	"github.com/SpectoLabs/hoverfly/core/util"
	"github.com/codegangsta/negroni"
	"github.com/go-zoo/bone"

	log "github.com/sirupsen/logrus"
)

type HoverflySimulation interface {
	GetSimulation() (SimulationViewV5, error)
	GetFilteredSimulation(string) (SimulationViewV5, error)
	PutSimulation(SimulationViewV5) SimulationImportResult
	DeleteSimulation()
}

type SimulationHandler struct {
	Hoverfly HoverflySimulation
}

func (this *SimulationHandler) RegisterRoutes(mux *bone.Mux, am *handlers.AuthHandler) {
	mux.Get("/api/v2/simulation", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.Get),
	))
	mux.Put("/api/v2/simulation", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.Put),
	))
	mux.Put("/api/v2/simulation/:id", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.PutWithId),
	))
	mux.Post("/api/v2/simulation", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.Post),
	))
	mux.Delete("/api/v2/simulation", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.Delete),
	))
	mux.Delete("/api/v2/simulation/:id", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.DeleteWithId),
	))
	mux.Options("/api/v2/simulation", negroni.New(
		negroni.HandlerFunc(this.Options),
	))

	mux.Get("/api/v2/simulation/schema", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.GetSchema),
	))
	mux.Options("/api/v2/simulation/schema", negroni.New(
		negroni.HandlerFunc(this.Options),
	))
}

func (this *SimulationHandler) Get(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	urlPattern := req.URL.Query().Get("urlPattern")

	var err error
	var simulationView SimulationViewV5
	if urlPattern == "" {
		simulationView, err = this.Hoverfly.GetSimulation()
	} else {
		simulationView, err = this.Hoverfly.GetFilteredSimulation(urlPattern)
	}
	if err != nil {
		handlers.WriteErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, _ := util.JSONMarshal(simulationView)

	handlers.WriteResponse(w, bytes)
}

func (this *SimulationHandler) Put(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	err := this.addSimulation(w, req, true)
	if err != nil {
		return
	}

	this.Get(w, req, next)
}

func (this *SimulationHandler) PutWithId(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	id := bone.GetValue(req, "id")
	fmt.Printf("%s\n", id)
	err := this.addSingleSimulation(w, req, true, id)
	if err != nil {
		return
	}

	this.Get(w, req, next)
}

func (this *SimulationHandler) Post(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	err := this.addSimulation(w, req, false)
	if err != nil {
		return
	}

	this.Get(w, req, next)
}

func (this *SimulationHandler) Delete(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	this.Hoverfly.DeleteSimulation()

	this.Get(w, req, next)
}

func (this *SimulationHandler) DeleteWithId(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	id := bone.GetValue(req, "id")

	simulationView, _ := this.Hoverfly.GetSimulation()

	var filtered []RequestMatcherResponsePairViewV5
	for _, value := range simulationView.DataViewV5.RequestResponsePairs {
		if value.Id == "" {
			filtered = append(filtered, value)
		} else if value.Id != id {
			filtered = append(filtered, value)
		}
	}
	simulationView.RequestResponsePairs = filtered
	this.Hoverfly.DeleteSimulation()

	result := this.Hoverfly.PutSimulation(simulationView)
	if result.Err != nil {
		handlers.WriteErrorResponse(w, "An error occurred: "+result.Err.Error(), http.StatusInternalServerError)
		return
	}
	if len(result.WarningMessages) > 0 {
		bytes, _ := util.JSONMarshal(result)

		handlers.WriteResponse(w, bytes)
		return
	}

	this.Get(w, req, next)
}

func (this *SimulationHandler) Options(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Add("Allow", "OPTIONS, GET, PUT, DELETE")
	handlers.WriteResponse(w, []byte(""))
}

func (this *SimulationHandler) GetSchema(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	handlers.WriteResponse(w, SimulationViewV5Schema)
}

func (this *SimulationHandler) OptionsSchema(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Add("Allow", "OPTIONS, GET")
}

func (this *SimulationHandler) addSimulation(w http.ResponseWriter, req *http.Request, overrideExisting bool) error {
	body, _ := ioutil.ReadAll(req.Body)

	simulationView, err := NewSimulationViewFromRequestBody(body)
	if err != nil {
		handlers.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if overrideExisting {
		this.Hoverfly.DeleteSimulation()
	}

	result := this.Hoverfly.PutSimulation(simulationView)
	if result.Err != nil {

		log.WithFields(log.Fields{
			"body": string(body),
		}).Debug(result.Err.Error())

		handlers.WriteErrorResponse(w, "An error occurred: "+result.Err.Error(), http.StatusInternalServerError)
		return result.Err
	}
	if len(result.WarningMessages) > 0 {
		bytes, _ := util.JSONMarshal(result)

		handlers.WriteResponse(w, bytes)
		return fmt.Errorf("import simulation result has warnings")
	}
	return nil
}

func (this *SimulationHandler) addSingleSimulation(w http.ResponseWriter, req *http.Request, overrideExisting bool, id string) error {
	body, _ := ioutil.ReadAll(req.Body)

	simulationSingleViewV1, err := NewSimulationSingleViewFromRequestBody(body)
	if err != nil {
		handlers.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return err
	}

	fmt.Printf("요청값 %+v\n\n\n\n", simulationSingleViewV1)

	simulationView, _ := this.Hoverfly.GetSimulation()

	var filtered []RequestMatcherResponsePairViewV5

	for _, value := range simulationView.DataViewV5.RequestResponsePairs {
		if value.Id == "" {
			filtered = append(filtered, value)
		} else if value.Id != id {
			filtered = append(filtered, value)
		}
	}
	simulationSingleViewV1.Id = id
	filtered = append(filtered, simulationSingleViewV1.RequestMatcherResponsePairViewV5)
	fmt.Printf("%+v|%+v\n", filtered, filtered[0].Id)
	simulationView.DataViewV5.RequestResponsePairs = filtered

	this.Hoverfly.DeleteSimulation()
	result := this.Hoverfly.PutSimulation(simulationView)
	if result.Err != nil {
		log.WithFields(log.Fields{
			"body": string(body),
		}).Debug(result.Err.Error())

		handlers.WriteErrorResponse(w, "An error occurred: "+result.Err.Error(), http.StatusInternalServerError)
		return result.Err
	}
	if len(result.WarningMessages) > 0 {
		bytes, _ := util.JSONMarshal(result)

		handlers.WriteResponse(w, bytes)
		return fmt.Errorf("import simulation result has warnings")
	}
	return nil
}
