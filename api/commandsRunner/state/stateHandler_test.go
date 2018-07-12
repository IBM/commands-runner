/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package state

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extension"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	yaml "gopkg.in/yaml.v2"
)

var statesJson string
var stateJson string

func init() {
	directorlog := "sample-director line 1\nsample-director line 2\nsample-director line 3\nsample-director line 4\nsample-director line 5"
	cflog := "sample-cf line 1\nsample-cf line 2\nsample-cf line 3\nsample-cf line 4\nsample-cf line 5"
	ioutil.WriteFile("/tmp/sample-director.log", []byte(directorlog), 0600)
	errDirector := os.Chmod("/tmp/sample-director.log", 0600)
	if errDirector != nil {
		log.Fatal(errDirector.Error())
	}
	ioutil.WriteFile("/tmp/sample-cf.log", []byte(cflog), 0600)
	errCF := os.Chmod("/tmp/sample-cf.log", 0600)
	if errCF != nil {
		log.Fatal(errCF.Error())
	}
	statesJson = "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	stateJson = "{\"name\":\"cfp-ext-template\",\"label\":\"Insert\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}"
}

func TestStatesOk(t *testing.T) {
	t.Log("Entering................. TestStatesOk")
	extension.SetExtensionPath("../../test/data/extensions/")
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/states", nil)
	//	addStateManagerToMap("TestStatesOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/states?extension-name=TestStatesOk", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var states States
	statesData, err := ioutil.ReadFile("../../test/resource/states.yaml")
	t.Logf("StatesData=%s", statesData)
	if err != nil {
		t.Fatal(err)
	}
	// Parse state file into the States structure
	err = yaml.Unmarshal(statesData, &states)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("States YAML:%s", statesData)
	expected, _ := json.MarshalIndent(states, "", "  ")
	got := strings.TrimRight(rr.Body.String(), "\n")
	t.Log(string(expected))
	t.Log(got)
	if got != string(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), string(expected))
	}
}

func TestStateOk(t *testing.T) {
	t.Log("Entering................. TestStateOk")
	SetStatePath("../../test/resource/states.yaml")
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	req, err := http.NewRequest("GET", "/cr/v1/state/director", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	var expected bytes.Buffer
	err = json.Indent(&expected, []byte(`{"name":"director","phase":"","label":"Director","log_path":"/tmp/sample-director.log","status":"READY","start_time":"","end_time":"","reason":"","script":"test","script_timeout":10,"protected":false,"deleted":false,"states_to_rerun":[],"previous_states":[],"next_states":["repeat"]}`), "", "  ")
	if err != nil {
		t.Error(err.Error())
	}
	got := strings.TrimRight(rr.Body.String(), "\n")
	t.Log(expected.String())
	t.Log(got)
	if got != expected.String() {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected.String())
	}
}

func TestInsertDeleteStateStates(t *testing.T) {
	t.Log("Entering................. TestInsertDeleteStateStates")
	stateFile := "../../test/resource/states-insert-delete.yaml"
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetStatePath(stateFile)
	addStateManagerToMap("cfp-ext-template", stateFile)
	inFileData, err := ioutil.ReadFile(stateFile)
	t.Log(string(inFileData))
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template&action=insert&pos=1&before=true", strings.NewReader(stateJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	middleFileData, err := ioutil.ReadFile(stateFile)
	t.Log(string(middleFileData))

	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template&action=delete&pos=1", strings.NewReader(stateJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	outFileData, err := ioutil.ReadFile(stateFile)

	if string(outFileData) != string(inFileData) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(outFileData), string(inFileData))
	}
}

func TestSetEmptyStates(t *testing.T) {
	t.Log("Entering................. TestInsertStateEmptyStates")
	stateFile := "../../test/resource/states-insert-delete.yaml"
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetStatePath(stateFile)
	addStateManagerToMap("cfp-ext-template", stateFile)
	inFileData, err := ioutil.ReadFile(stateFile)
	t.Log(string(inFileData))
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

}

func TestInsertDeleteStateStatesAutoLocation(t *testing.T) {
	t.Log("Entering................. TestInsertDeleteStateStatesAutoLocation")
	log.SetLevel(log.DebugLevel)
	stateFile := "../../test/resource/states-insert-delete-auto-location.yaml"
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetStatePath(stateFile)
	addStateManagerToMap("cfp-ext-template-auto-location", stateFile)
	inFileData, err := ioutil.ReadFile(stateFile)
	t.Log(string(inFileData))
	//stateAutoLocationJson := "{\"name\":\"cfp-ext-template-auto-location\",\"label\":\"Insert\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}"
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template-auto-location&insert-extension-name=cfp-ext-template-auto-location&action=insert&pos=0", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Log(rr.Body)
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	middleFileData, err := ioutil.ReadFile(stateFile)
	t.Log(string(middleFileData))

	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template-auto-location&action=delete&state-name=cfp-ext-template-auto-location&pos=0", strings.NewReader(stateJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Log(rr.Body)
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	outFileData, err := ioutil.ReadFile(stateFile)

	if string(outFileData) != string(inFileData) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(outFileData), string(inFileData))
	}
}

func TestInsertDeleteStateStatesByName(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestInsertDeleteStateStatesByName")
	stateFile := "../../test/resource/states-insert-delete.yaml"
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetStatePath(stateFile)
	addStateManagerToMap("cfp-ext-template", stateFile)
	inFileData, err := ioutil.ReadFile(stateFile)
	t.Log("inFileData:\n" + string(inFileData))
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template&action=insert&pos=0&before=true&state-name=task1", strings.NewReader(stateJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	middleFileData, err := ioutil.ReadFile(stateFile)
	t.Log("middleFileData:\n" + string(middleFileData))

	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=cfp-ext-template&action=delete&pos=0&state-name=cfp-ext-template", strings.NewReader(stateJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	outFileData, err := ioutil.ReadFile(stateFile)

	if string(outFileData) != string(inFileData) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(outFileData), string(inFileData))
	}
}

/* Remove as to risky
func TestStatesPutJson(t *testing.T) {
	t.Log("Entering................. TestStatesPutJson")
	addStateManagerToMap("TestStatesPutJson", "../../test/resource/states-post-sample-from-json.yaml")
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=TestStatesPutJson", strings.NewReader(statesJson))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}

}

func TestStatesPutYaml(t *testing.T) {
	t.Log("Entering................. TestStatesPutYaml")
	addStateManagerToMap("TestStatesPutYaml", "../../test/resource/states-post-sample-from-yaml.yaml")
	statesYaml, _ := ioutil.ReadFile("../../test/resource/states-post-sample.yaml")
	t.Logf("statesYaml:\n%s", string(statesYaml))
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=TestStatesPutYaml", bytes.NewReader(statesYaml))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStates)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}

	statesYamlOut, _ := ioutil.ReadFile("../../test/resource/states-post-sample-from-yaml.yaml")
	if bytes.Compare(statesYaml, statesYamlOut) != 0 {
		t.Errorf("Output file not the same as input")
	}
}
*/
func TestStateNotMethodDELETE(t *testing.T) {
	t.Log("Entering................. TestStateNotMethodDELETE")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateNotMethodDELETE", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("DELETE", "/cr/v1/state/director?extension-name=TestStateNotMethodDELETE", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("DELETE", "/cr/v1/state/director", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestStateNOk(t *testing.T) {
	t.Log("Entering................. TestStateNOk")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateNOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists?extension-name=TestStateNOk", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestPutState(t *testing.T) {
	t.Log("Entering................. TestPutState")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestPutState", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("PUT", "/cr/v1/state/director?status=TEST&extension-name=TestPutState", nil)
	extension.SetExtensionPath("../../test/data/extensions/")
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("PUT", "/cr/v1/state/director?status=FAILED", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err = http.NewRequest("GET", "/cr/v1/state/director", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	//var expected bytes.Buffer
	//json.Indent(&expected, []byte(`{"name":"director","label":"Director","status":"TEST","start_time":"","end_time":"","reason":""}`), "", "  ")
	//got := strings.TrimRight(rr.Body.String(), "\n")
	var stateAux State
	json.Unmarshal(rr.Body.Bytes(), &stateAux)
	if stateAux.Status != "FAILED" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			stateAux.Status, "FAILED")
	}

	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

}

func TestStatusNotMethodPOST(t *testing.T) {
	t.Log("Entering................. TestStatusNotMethodPOST")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStatusNotMethodPOST", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("POST", "/cr/v1/state/director?status=TEST&extension-name=TestStatusNotMethodPOST", nil)
	extension.SetExtensionPath("../../test/data/extensions/")
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("POST", "/cr/v1/state/director?status=TEST", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestStatusNotMethodDELETE(t *testing.T) {
	t.Log("Entering................. TestStatusNotMethodDELETE")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStatusNotMethodDELETE", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("DELETE", "/cr/v1/state/director?status=TEST&extension-name=TestStatusNotMethodDELETE", nil)
	extension.SetExtensionPath("../../test/data/extensions/")
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("DELETE", "/cr/v1/state/director?status=TEST", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestInvalidSubcommand(t *testing.T) {
	t.Log("Entering................. TestInvalidSubcommand")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestInvalidSubcommand", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/director/invalidsubcmd?status=TEST&extension-name=TestInvalidSubcommand", nil)
	extension.SetExtensionPath("../../test/data/extensions/")
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/director/invalidsubcmd?status=TEST", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestAddStateManagerIBM(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestAddStateManagerIBM")
	extension.SetExtensionPath("../../test/data/extensions/")
	extension.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	global.SetExtensionResourcePath("../../test/resource/extensions/")
	extension := "cfp-ext-template"
	err := addStateManager(extension)
	if err != nil {
		t.Error("Unable to add state manager")
	}
	sm, err := getStateManager(extension)
	if err != nil {
		t.Error("Unable to retrieve state manager " + extension)
	}
	expected := "../../test/data/extensions/embedded/" + extension + "/statesFile-" + extension + ".yml"
	got := sm.StatesPath
	if expected != got {
		t.Error("Expecting " + expected + " and got " + got)
	}
}

func TestStateLogOk(t *testing.T) {
	t.Log("Entering................. TestStateLogOk")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?extension-name=TestStateLogOk", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/director/log", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	t.Log(rr.Body.String())
}

func TestStateLogNOk(t *testing.T) {
	t.Log("Entering................. TestStateLogNOk")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogNOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNOk", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists/log", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestStateLogNotMethodPUT(t *testing.T) {
	t.Log("Entering................. TestStateLogNotMethodPUT")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogNotMethodPUT", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("PUT", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodPUT", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("PUT", "/cr/v1/state/not-exists/log", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestStateLogNotMethodPOST(t *testing.T) {
	t.Log("Entering................. TestStateLogNotMethodPOST")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogNotMethodPOST", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("POST", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodPOST", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("POST", "/cr/v1/state/not-exists/log", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestStateLogNotMethodDELETE(t *testing.T) {
	t.Log("Entering................. TestStateLogNotMethodDELETE")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogNotMethodDELETE", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("DELETE", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodDELETE", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("DELETE", "/cr/v1/state/not-exists/log", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestStateLogFromToOk(t *testing.T) {
	t.Log("Entering................. TestStateLogFromToOk")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogFromToOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?first-line=2&length=2&extension-name=TestStateLogFromToOk", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?first-line=2&length=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	logData := string(rr.Body.String())
	t.Log(logData)
	if (strings.Contains(logData, "sample-director line 1") ||
		strings.Contains(logData, "sample-director line 4") ||
		strings.Contains(logData, "sample-director line 5")) &&
		!strings.Contains(logData, "sample-director line 3") &&
		!strings.Contains(logData, "sample-director line 4") {
		t.Error("Expecting 'director' in reponse but gets:" + logData)
	}
	t.Log("Exiting................. TestStateLogFromToOk")
}

func TestStateLogFromNotInteger(t *testing.T) {
	t.Log("Entering................. TestStateLogFromNotInteger")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogFromNotInteger", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?first-line=a&extension-name=TestStateLogFromNotInteger", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?first-line=a", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestStateLogToNotInteger(t *testing.T) {
	t.Log("Entering................. TestStateLogToNotInteger")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManagerToMap("TestStateLogToNotInteger", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?length=v&extension-name=TestStateLogToNotInteger", nil)
	SetStatePath("../../test/resource/states.yaml")
	req, err := http.NewRequest("GET", "/cr/v1/state/director/log?length=v", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleState)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
