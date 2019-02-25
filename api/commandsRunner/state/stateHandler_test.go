/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
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
	"path/filepath"
	"strings"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var statesJson string
var stateJson string
var stateInsertDeleteJson string

func init() {
	state1log := "sample-state1 line 1\nsample-state1 line 2\nsample-state1 line 3\nsample-state1 line 4\nsample-state1 line 5"
	state2log := "sample-state2 line 1\nsample-state2 line 2\nsample-state2 line 3\nsample-state2 line 4\nsample-state2 line 5"
	ioutil.WriteFile("/tmp/sample-state1.log", []byte(state1log), 0600)
	errstate1 := os.Chmod("/tmp/sample-state1.log", 0600)
	if errstate1 != nil {
		log.Fatal(errstate1.Error())
	}
	ioutil.WriteFile("/tmp/sample-state2.log", []byte(state2log), 0600)
	errstate2 := os.Chmod("/tmp/sample-state2.log", 0600)
	if errstate2 != nil {
		log.Fatal(errstate2.Error())
	}
	statesJson = "{\"states\":[{\"name\":\"state1\",\"label\":\"Step 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runer\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	stateJson = "{\"name\":\"ext-template\",\"label\":\"Insert\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}"
	stateInsertDeleteJson = "{\"name\":\"ext-template-insert-delete-handler\",\"label\":\"Insert\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}"
}

func TestStatesOk(t *testing.T) {
	t.Log("Entering................. TestStatesOk")
	extensionPath, err := global.CopyToTemp("TestStatesOk", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/states?extension-name=state-handler-test", nil)
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
	global.RemoveTemp("TestStatesOk")
}

func TestStateOk(t *testing.T) {
	t.Log("Entering................. TestStateOk")
	extensionPath, err := global.CopyToTemp("TestStateOk", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	req, err := http.NewRequest("GET", "/cr/v1/state/state1?extension-name=state-handler-test", nil)
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
	err = json.Indent(&expected, []byte(`{"name":"state1","phase":"","label":"State 1","log_path":"/tmp/sample-state1.log","status":"READY","start_time":"","end_time":"","reason":"","script":"test","script_timeout":10,"protected":false,"deleted":false, "prerequisite_states":[],"states_to_rerun":[],"rerun_on_run_of_states":[],"previous_states":[],"next_states":["repeat"],"executed_by_extension_name": "","execution_id": 0,"next_run": false}`), "", "  ")
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
	global.RemoveTemp("TestStateOk")
}

func TestInsertDeleteStateStates(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestInsertDeleteStateStates")
	inFileData, err := ioutil.ReadFile("../../test/resource/ext-insert-delete-states-file-handler.yml")
	t.Log(string(inFileData))
	extensionPath, err := global.CopyToTemp("TestInsertDeleteStateStates", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	stateFile := filepath.Join(extensionPath, "embedded/ext-insert-delete-handler/states-file.yml")
	ioutil.WriteFile(stateFile, inFileData, 0644)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	addStateManager("ext-insert-delete")
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-handler&action=insert&pos=1&before=true", strings.NewReader(stateInsertDeleteJson))
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

	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-handler&action=delete&pos=1", strings.NewReader(stateJson))
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
	global.RemoveTemp("TestInsertDeleteStateStates")
}

func TestSetEmptyStates(t *testing.T) {
	t.Log("Entering................. TestSetEmptyStates")
	inFileData, err := ioutil.ReadFile("../../test/resource/ext-insert-delete-states-file.yml")
	t.Log(string(inFileData))
	extensionPath, err := global.CopyToTemp("TestSetEmptyStates", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	stateFile := filepath.Join(extensionPath, "embedded/ext-insert-delete/states-file.yml")
	ioutil.WriteFile(stateFile, inFileData, 0644)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	addStateManager("ext-insert-delete")
	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-template", strings.NewReader(""))
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
	global.RemoveTemp("TestSetEmptyStates")

}

func TestInsertDeleteStateStatesAutoLocation(t *testing.T) {
	t.Log("Entering................. TestInsertDeleteStateStatesAutoLocation")
	//log.SetLevel(log.DebugLevel)
	inFileData, err := ioutil.ReadFile("../../test/resource/ext-insert-delete-states-file.yml")
	t.Log("inFileData:\n" + string(inFileData))
	extensionPath, err := global.CopyToTemp("TestInsertDeleteStateStatesAutoLocation", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	stateFile := filepath.Join(extensionPath, "embedded/ext-insert-delete-auto-location/states-file.yml")
	ioutil.WriteFile(stateFile, inFileData, 0644)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	addStateManager("ext-insert-delete-auto-location")

	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-auto-location&insert-extension-name=ext-template-auto&state-name=task1&action=insert&pos=0", nil)
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
	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-auto-location&action=delete&state-name=ext-template-auto&pos=0", strings.NewReader(stateJson))
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
	global.RemoveTemp("TestInsertDeleteStateStatesAutoLocation")
}

func TestInsertDeleteStateStatesByName(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestInsertDeleteStateStatesByName")
	inFileData, err := ioutil.ReadFile("../../test/resource/ext-insert-delete-states-file.yml")
	t.Log("inFileData:\n" + string(inFileData))
	extensionPath, err := global.CopyToTemp("TestInsertDeleteStateStatesByName", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	stateFile := filepath.Join(extensionPath, "embedded/ext-insert-delete-by-name/states-file.yml")
	ioutil.WriteFile(stateFile, inFileData, 0644)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	addStateManager("ext-insert-delete-by-name")

	req, err := http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-by-name&action=insert&pos=0&before=true&state-name=task1", strings.NewReader(stateInsertDeleteJson))
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
	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=ext-insert-delete-by-name&action=delete&pos=0&state-name=ext-template-insert-delete-handler", strings.NewReader(stateInsertDeleteJson))
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
	global.RemoveTemp("TestInsertDeleteStateStatesByName")
}

func TestStateNotMethodDELETE(t *testing.T) {
	t.Log("Entering................. TestStateNotMethodDELETE")
	req, err := http.NewRequest("DELETE", "/cr/v1/state/state1", nil)
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
	extensionPath, err := global.CopyToTemp("TestStateNOk", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists?extension-name=ext-template", nil)
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
	global.RemoveTemp("TestStateNOk")
}

func TestPutState(t *testing.T) {
	t.Log("Entering................. TestPutState")
	extensionPath, err := global.CopyToTemp("TestPutState", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("PUT", "/cr/v1/state/state1?extension-name=state-handler-reset&status=FAILED", nil)
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
	req, err = http.NewRequest("GET", "/cr/v1/state/state1?extension-name=state-handler-reset", nil)
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

	var stateAux State
	json.Unmarshal(rr.Body.Bytes(), &stateAux)
	if stateAux.Status != "FAILED" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			stateAux.Status, "FAILED")
	}

	req, err = http.NewRequest("PUT", "/cr/v1/states?extension-name=state-handler-reset&action=reset", nil)
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
	global.RemoveTemp("TestPutState")

}

func TestStatusNotMethodPOST(t *testing.T) {
	t.Log("Entering................. TestStatusNotMethodPOST")
	extensionPath, err := global.CopyToTemp("TestStatusNotMethodPOST", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("POST", "/cr/v1/state/state1?extension-name=ext-template&status=TEST", nil)
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
	global.RemoveTemp("TestStatusNotMethodPOST")
}

func TestStatusNotMethodDELETE(t *testing.T) {
	t.Log("Entering................. TestStatusNotMethodDELETE")
	extensionPath, err := global.CopyToTemp("TestStatusNotMethodDELETE", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("DELETE", "/cr/v1/state/state1?extension-name=ext-template&status=TEST", nil)
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
	global.RemoveTemp("TestStatusNotMethodDELETE")
}

func TestInvalidSubcommand(t *testing.T) {
	t.Log("Entering................. TestInvalidSubcommand")
	extensionPath, err := global.CopyToTemp("TestInvalidSubcommand", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/state1/invalidsubcmd?extension-name=ext-template&status=TEST", nil)
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
	global.RemoveTemp("TestInvalidSubcommand")
}

func TestAddStateManagerIBM(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestAddStateManagerIBM")
	extensionPath, err := global.CopyToTemp("TestAddStateManagerIBM", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	//	global.SetExtensionResourcePath("../../test/resource/extensions/")
	extension := "ext-template"
	addStateManager(extension)
	sm, err := getStatesManager(extension)
	if err != nil {
		t.Error("Unable to retrieve state manager " + extension)
	}
	expected := filepath.Join(extensionPath, "embedded", extension, "states-file.yml")
	got := sm.StatesPath
	if expected != got {
		t.Error("Expecting " + expected + " and got " + got)
	}
	global.RemoveTemp("TestAddStateManagerIBM")
}

func TestStateLogOk(t *testing.T) {
	t.Log("Entering................. TestStateLogOk")
	extensionPath, err := global.CopyToTemp("TestStateLogOk", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/state1/log?extension-name=state-handler-log", nil)
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
	global.RemoveTemp("TestStateLogOk")
}

func TestStateLogNOk(t *testing.T) {
	t.Log("Entering................. TestStateLogNOk")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	//	addStateManager("TestStateLogNOk", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNOk", nil)
	req, err := http.NewRequest("GET", "/cr/v1/state/not-exists/log?extension-name=ext-template", nil)
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
	//	addStateManager("TestStateLogNotMethodPUT", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("PUT", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodPUT", nil)
	req, err := http.NewRequest("PUT", "/cr/v1/state/not-exists/log?extension-name=ext-template", nil)
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
	//	addStateManager("TestStateLogNotMethodPOST", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("POST", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodPOST", nil)
	req, err := http.NewRequest("POST", "/cr/v1/state/not-exists/log?extension-name=ext-template", nil)
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
	//	addStateManager("TestStateLogNotMethodDELETE", "../../test/resource/states.yaml")
	//	req, err := http.NewRequest("DELETE", "/cr/v1/state/not-exists/log?extension-name=TestStateLogNotMethodDELETE", nil)
	req, err := http.NewRequest("DELETE", "/cr/v1/state/not-exists/log?extension-name=ext-template", nil)
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
	extensionPath, err := global.CopyToTemp("TestStateLogFromToOk", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/state1/log?extension-name=state-handler-log&first-line=2&length=2", nil)
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
	if (strings.Contains(logData, "sample-state1 line 1") ||
		strings.Contains(logData, "sample-state1 line 4") ||
		strings.Contains(logData, "sample-state1 line 5")) &&
		!strings.Contains(logData, "sample-state1 line 3") &&
		!strings.Contains(logData, "sample-state1 line 4") {
		t.Error("Expecting 'state1' in reponse but gets:" + logData)
	}
	t.Log("Exiting................. TestStateLogFromToOk")
	global.RemoveTemp("TestStateLogFromToOk")
}

func TestStateLogFromNotInteger(t *testing.T) {
	t.Log("Entering................. TestStateLogFromNotInteger")
	extensionPath, err := global.CopyToTemp("TestStateLogFromNotInteger", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/state1/log?extension-name=state-handler-log&first-line=a", nil)
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
	global.RemoveTemp("TestStateLogFromNotInteger")
}

func TestStateLogToNotInteger(t *testing.T) {
	t.Log("Entering................. TestStateLogToNotInteger")
	extensionPath, err := global.CopyToTemp("TestStateLogToNotInteger", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/state/state1/log?extension-name=state-handler-log&length=v", nil)
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
	global.RemoveTemp("TestStateLogToNotInteger")
}
