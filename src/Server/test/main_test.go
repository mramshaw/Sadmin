package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	// local import
	"admin-server/application"
)

var app application.App

var authUser, authPassword string

func TestMain(m *testing.M) {
	authUser = os.Getenv("AUTH_USER")
	authPassword = os.Getenv("AUTH_PASSWORD")
	app = application.App{}
	app.Initialize(
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_DB"),
		authUser,
		authPassword)
	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func TestEmptyTables(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/servers", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetBadServer(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/servers/a", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetNonExistentServer(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/servers/11", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Server not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Server not found'. Got '%s'", m["error"])
	}
}

func TestCreateServerNoCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test server"}`)

	req, err := http.NewRequest("POST", "/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestCreateServerWithCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test server"}`)

	req, err := http.NewRequest("POST", "/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test server" {
		t.Errorf("Expected server name to be 'test server'. Got '%v'", m["name"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected server ID to be '1'. Got '%v'", m["id"])
	}
}

func TestCreateDuplicateServerWithCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test server"}`)

	req, err := http.NewRequest("POST", "/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	// Now check duplicate

	req, err = http.NewRequest("POST", "/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on 2nd http.NewRequest: %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusConflict, response.Code)
}

func TestGetServer(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetServers(t *testing.T) {
	clearTables()
	addServers(3)

	req, err := http.NewRequest("GET", "/v1/servers?count=55&start=-1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)

	if len(mm) != 3 {
		t.Errorf("Expected '3' servers. Got '%v'", len(mm))
	}
}

func TestUpdatePutServerNoCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalServer map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalServer)

	payload := []byte(`{"name":"test server - updated"}`)

	req, err = http.NewRequest("PUT", "/v1/servers/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PUT): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestUpdatePutServerWithCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalServer map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalServer)

	payload := []byte(`{"name":"test server - updated"}`)

	req, err = http.NewRequest("PUT", "/v1/servers/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PUT): %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalServer["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalServer["id"], m["id"])
	}
	if m["name"] == originalServer["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalServer["name"], m["name"], m["name"])
	}
}

func TestUpdatePatchServerNoCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalServer map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalServer)

	payload := []byte(`{"name":"test server - updated"}`)

	req, err = http.NewRequest("PATCH", "/v1/servers/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PATCH): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestUpdatePatchServerWithCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalServer map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalServer)

	payload := []byte(`{"name":"test server - updated"}`)

	req, err = http.NewRequest("PATCH", "/v1/servers/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PATCH): %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalServer["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalServer["id"], m["id"])
	}
	if m["name"] == originalServer["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalServer["name"], m["name"], m["name"])
	}
}

func TestDeleteServerNoCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (DELETE): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestDeleteServerWithCredentials(t *testing.T) {
	clearTables()
	addServers(1)

	req, err := http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (DELETE): %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("GET", "/v1/servers/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (Second GET): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTablesExist() {
	if _, err := app.DB.Exec(serversTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	app.DB.Exec("DELETE FROM servers")
	app.DB.Exec("ALTER TABLE servers AUTO_INCREMENT = 1")
}

func TestSearch(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test server"}`)

	req, err := http.NewRequest("POST", "/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (1st POST): %s", err)
	}
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var bb bytes.Buffer
	mw := multipart.NewWriter(&bb)
	mw.WriteField("count", "1")
	mw.WriteField("start", "0")
	mw.WriteField("name", "won't match")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (2nd POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)

	if len(mm) != 0 {
		t.Errorf("2nd Post - Expected no matching servers. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "1")
	mw.WriteField("start", "0")
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (3rd POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)
	// only want the first one
	m := mm[0]

	if m["name"] != "test server" {
		t.Errorf("3rd Post - Expected server name to be 'test server'. Got '%v'", m["name"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("3rd Post - Expected server ID to be '1'. Got '%v'", m["id"])
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "15")
	mw.WriteField("start", "-5")
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (4th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)
	// only want the first one
	m = mm[0]

	if m["name"] != "test server" {
		t.Errorf("4th Post - Expected server name to be 'test server'. Got '%v'", m["name"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("4th Post - Expected server ID to be '1'. Got '%v'", m["id"])
	}

	addServers(12)

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (5th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 10 {
		t.Errorf("5th Post - Expected '10' servers. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "5")
	mw.WriteField("start", "3")
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (6th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 5 {
		t.Errorf("6th Post - Expected '5' servers. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "-3") // Should reset to 25
	mw.WriteField("start", "-5") // Should reset to 0
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (7th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 13 {
		t.Errorf("7th Post - Expected '13' servers. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "50") // Should reset to 25
	mw.WriteField("start", "0")
	mw.WriteField("name", "%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (8th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 13 {
		t.Errorf("8th Post - Expected '13' servers. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "50") // Should reset to 25
	mw.WriteField("start", "0")
	mw.WriteField("name", "Server 1%")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/servers", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (9th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 4 {
		t.Errorf("9th Post - Expected '4' servers. Got '%v'", len(mm))
	}
}

func addServers(count int) {
	if count < 1 {
		count = 1
	}
	for i := 1; i < count+1; i++ {
		app.DB.Exec("INSERT INTO servers(name) VALUES(?)",
			"Server "+strconv.Itoa(i))
	}
}

const serversTableCreationQuery = `CREATE TABLE IF NOT EXISTS servers
(
	id BIGINT(20) AUTO_INCREMENT,
	name VARCHAR(50) NOT NULL UNIQUE,
	PRIMARY KEY (id)
)`
