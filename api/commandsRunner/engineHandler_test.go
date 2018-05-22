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
package commandsRunner

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func TestEngineStartPUT(t *testing.T) {
	t.Log("Entering................. TestEngineStartPUT")
	addStateManagerToMap("TestEngineStartPUT", "../test/resource/engine-run-success.yaml")
	extensionManager.SetExtensionPath("../test/data/extensions/")
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/ibm-test-extensions.txt")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	time.Sleep(10 * time.Second)

	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, err = http.NewRequest("GET", "/cr/v1/engine?extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestEngineStartExtensonPUT(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestEngineStartExtensonPUT")
	extensionManager.SetExtensionPath("../test/data/extensions/")
	global.SetExtensionResourcePath("api/test/resource/extensions/")
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/ibm-test-extensions.txt")
	extension := "cfp-ext-template"
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	time.Sleep(10 * time.Second)

	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, err = http.NewRequest("GET", "/cr/v1/engine?extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestEnginePUTRunning(t *testing.T) {
	t.Log("Entering................. TestEnginePUTRunning")
	addStateManagerToMap("TestEnginePUTRunning", "../test/resource/engine-run-running.yaml")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name=TestEnginePUTRunning", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusConflict)
	}

	req, err = http.NewRequest("GET", "/cr/v1/engine?extension-name=TestEnginePUTRunning", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusProcessing {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusProcessing)
	}

}

func TestEngineResetRunning(t *testing.T) {
	t.Log("Entering................. TestEngineResetRunning")
	addStateManagerToMap("TestEngineResetRunning", "../test/resource/states-reset.yaml")

	rr := httptest.NewRecorder()

	req, errState := http.NewRequest("PUT", "/cr/v1/state/task1?status=RUNNING&extension-name=TestEngineResetRunning", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus := http.HandlerFunc(handleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not set task1 as RUNNING")
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name=TestEngineResetRunning", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Log(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestEngineReset(t *testing.T) {
	t.Log("Entering................. TestEngineReset")
	addStateManagerToMap("TestEngineReset", "../test/resource/states-reset.yaml")

	rr := httptest.NewRecorder()

	req, errState := http.NewRequest("PUT", "/cr/v1/state/task1?status=FAILED&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus := http.HandlerFunc(handleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, errState = http.NewRequest("PUT", "/cr/v1/state/task2?status=SKIP&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus = http.HandlerFunc(handleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, errState = http.NewRequest("PUT", "/cr/v1/state/task3?status=READY&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus = http.HandlerFunc(handleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name=TestEngineReset", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(handleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Log(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
