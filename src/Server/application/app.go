// Package application is a package encompassing the bulk of the application.
package application

import (
	// native packages
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	// local packages
	"admin-server/servers"

	// GitHub packages
	"github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

// App represents the application
type App struct {
	Router *httprouter.Router
	DB     *sql.DB
}

func (a *App) getServerEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid server ID")
		return
	}
	s := servers.Server{ID: int64(id)}
	if err := s.GetServer(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Server not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) getServersEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	if count > 25 || count < 1 {
		count = 25
	}
	if start < 0 {
		start = 0
	}
	servers, err := servers.GetServers(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, servers)
}

func (a *App) createServerEndpoint(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var s servers.Server
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := s.CreateServer(a.DB); err != nil {
		merr, ok := err.(*mysql.MySQLError)
		if !ok {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Check for Duplicate
		if merr.Number == 1062 {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, s)
}

func (a *App) modifyServerEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid server ID")
		return
	}
	var s servers.Server
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	s.ID = int64(id)
	if _, err := s.UpdateServer(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) deleteServerEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid server ID")
		return
	}
	s := servers.Server{ID: int64(id)}
	if _, err := s.DeleteServer(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) searchServersEndpoint(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))
	name := req.FormValue("name")

	if count > 25 || count < 1 {
		count = 25
	}
	if start < 0 {
		start = 0
	}

	servers, err := servers.SearchServers(a.DB, start, count, name)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, servers)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}

func basicAuth(h httprouter.Handle, requiredUser, requiredPassword string) httprouter.Handle {

	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Get the Basic Authentication credentials
		user, password, hasAuth := req.BasicAuth()

		if hasAuth && user == requiredUser && password == requiredPassword {
			// Delegate request to the given handle
			h(w, req, ps)
		} else {
			// Request Basic Authentication otherwise
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

// Initialize sets up the database connection, router, and routes for the app
func (a *App) Initialize(dbHost, dbPort, dbUser, dbPassword, dbName, authUser, authPassword string) {

	// For SSL, specify '?tls=skip-verify'. For TLS, specify '?tls=true'.
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify", dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error

	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = httprouter.New()

	a.Router.GET("/v1/servers", a.getServersEndpoint)
	a.Router.POST("/v1/servers", basicAuth(a.createServerEndpoint, authUser, authPassword))
	a.Router.GET("/v1/servers/:id", a.getServerEndpoint)
	a.Router.PUT("/v1/servers/:id", basicAuth(a.modifyServerEndpoint, authUser, authPassword))
	a.Router.PATCH("/v1/servers/:id", basicAuth(a.modifyServerEndpoint, authUser, authPassword))
	a.Router.DELETE("/v1/servers/:id", basicAuth(a.deleteServerEndpoint, authUser, authPassword))
	a.Router.POST("/v1/search/servers", a.searchServersEndpoint)
}

// Run starts the app and serves on the specified port
func (a *App) Run(port string) {
	log.Print("Now serving servers ...")
	log.Fatal(http.ListenAndServeTLS(":"+port, "../../certificates/REST-server.pem", "../../certificates/REST-server-private-key.pem", a.Router))
}
