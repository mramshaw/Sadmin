package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

const httpTimeout int = 3 // seconds

var timeout = time.Duration(time.Duration(httpTimeout) * time.Second)

// ---------------------------------------
// According to the net/http documentation:
//     "Clients and Transports are safe for concurrent use by multiple goroutines
//      and for efficiency should only be created once and re-used."
var client = &http.Client{
	Timeout:   timeout,
	Transport: tr,
}
var tr = &http.Transport{
	// Disable certificate check, effectively dropping down to SSL
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// ---------------------------------------

type server struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var pageTemplates = template.Must(template.ParseGlob("../../templates/*.gohtml"))

var remoteHost string
var remotePort string
var remoteAuthUser string
var remoteAuthPass string

var authUser string
var authPass string

func listServersHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	req, err := http.NewRequest("GET", "https://"+remoteHost+":"+remotePort+"/v1/servers", nil)
	if err != nil {
		log.Printf("listServersHandler - Error on http.NewRequest: '%s'", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("listServersHandler - Error on request: '%v'", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("listServersHandler - Error on reading: '%v'", err)
		return
	}

	var serverList []server
	err = json.Unmarshal(body, &serverList)
	if err != nil {
		log.Printf("listServersHandler - Unmarshal error: '%v'", err)
	}

	log.Println("Listed Server Entries", resp.StatusCode)

	pageTemplates.ExecuteTemplate(w, "serverList.gohtml", serverList)
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

func main() {

	remoteHost = os.Getenv("REMOTE_HOST")
	remotePort = os.Getenv("REMOTE_PORT")
	remoteAuthUser = os.Getenv("REMOTE_AUTH_USER")
	remoteAuthPass = os.Getenv("REMOTE_AUTH_PASSWORD")

	port := os.Getenv("PORT")

	authUser = os.Getenv("AUTH_USER")
	authPass = os.Getenv("AUTH_PASSWORD")

	router := httprouter.New()

	// handle static assets (not logged)
	router.ServeFiles("/static/*filepath", http.Dir("../../assets"))

	router.GET("/Servers", basicAuth(listServersHandler, authUser, authPass))
	router.GET("/createServer", basicAuth(showCreateServerForm, authUser, authPass))
	router.POST("/createServer", basicAuth(createServerEntry, authUser, authPass))
	router.GET("/deleteServer", basicAuth(showDeleteServerForm, authUser, authPass))
	router.POST("/deleteServer", basicAuth(deleteServerEntry, authUser, authPass))

	log.Println("Now serving servers ...")
	log.Fatal(http.ListenAndServeTLS(":"+port, "../../certificates/WEB-server.pem", "../../certificates/WEB-server-private-key.pem", router))
}

type createPageVars struct {
	Name        string
	Invalid     bool
	Duplicate   bool
	Error       bool
	ErrorString string
}

func showCreateServerForm(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {
	page := createPageVars{Name: ""} // "required"}
	pageTemplates.ExecuteTemplate(writer, "createServer.gohtml", page)
}

func createServerEntry(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {

	page := createPageVars{request.FormValue("name"), false, false, false, ""}

	// Check for valid server name
	if !serverNameValid(request.FormValue("name")) {
		page.Invalid = true
		pageTemplates.ExecuteTemplate(writer, "createServer.gohtml", page)
		return
	}

	payload := []byte(`{"name":"` + request.FormValue("name") + `"}`)

	req, err := http.NewRequest("POST", "https://"+remoteHost+":"+remotePort+"/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("createServerEntry - Error on http.NewRequest: %s", err)
		return
	}
	req.SetBasicAuth(remoteAuthUser, remoteAuthPass)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("createServerEntry - Error on request: '%v'", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("createServerEntry - Error on reading: '%v'", err)
		return
	}

	// Check for duplicate
	if resp.StatusCode == http.StatusConflict {
		page.Duplicate = true
		pageTemplates.ExecuteTemplate(writer, "createServer.gohtml", page)
		return
	}

	// Check for errors
	if resp.StatusCode != http.StatusCreated {
		page.Error = true
		page.ErrorString = string(body)
		pageTemplates.ExecuteTemplate(writer, "createServer.gohtml", page)
		return
	}

	log.Println("Created Server Entry", resp.StatusCode, request.FormValue("name"))

	// redisplay servers list
	listServersHandler(writer, request, ps)
}

type deletePageVars struct {
	ID             int
	Name           string
	NoLongerExists bool
	Error          bool
	ErrorString    string
}

func showDeleteServerForm(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {
	id, _ := strconv.Atoi(request.FormValue("id"))
	page := deletePageVars{ID: id, Name: request.FormValue("name")}
	pageTemplates.ExecuteTemplate(writer, "deleteServer.gohtml", page)
}

func deleteServerEntry(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {

	id, _ := strconv.Atoi(request.FormValue("id"))
	page := deletePageVars{ID: id, Name: request.FormValue("name")}

	req, err := http.NewRequest("DELETE", "https://"+remoteHost+":"+remotePort+"/v1/servers/"+request.FormValue("id"), nil)
	if err != nil {
		log.Printf("deleteServerEntry - Error on http.NewRequest: %s", err)
		return
	}
	req.SetBasicAuth(remoteAuthUser, remoteAuthPass)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("deleteServerEntry - Error on request: '%v'", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("deleteServerEntry - Error on reading: '%v'", err)
		return
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		page.Error = true
		page.ErrorString = string(body)
		pageTemplates.ExecuteTemplate(writer, "deleteServer.gohtml", page)
		return
	}

	log.Println("Deleted Server Entry", resp.StatusCode, request.FormValue("name"))

	// redisplay servers list
	listServersHandler(writer, request, ps)
}
