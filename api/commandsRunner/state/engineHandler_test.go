/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
*/
package state

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	//	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func TestEngineStartPUT(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestEngineStartPUT")
	// addStateManager("TestEngineStartPUT", "../../test/resource/engine-run-success.yaml")
	addStateManager("TestEngineStartPUT")
	extensionPath, err := global.CopyToTemp("TestEngineStartPUT", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleEngine)

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
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not reset engine")
	}

	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset-execution-info&extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not reset execution info")
	}

	req, err = http.NewRequest("GET", "/cr/v1/engine?extension-name=TestEngineStartPUT", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	global.RemoveTemp("TestEngineStartPUT")
}

func TestEngineStartExtensonPUT(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestEngineStartExtensonPUT")
	extensionPath, err := global.CopyToTemp("TestEngineStartExtensonPUT", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extension := "ext-template"
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleEngine)

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
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not reset engine")
	}

	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset-execution-info&extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not reset execution info")
	}

	req, err = http.NewRequest("GET", "/cr/v1/engine?extension-name="+extension, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	global.RemoveTemp("TestEngineStartExtensonPUT")
}

func TestEnginePUTRunning(t *testing.T) {
	//	log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestEnginePUTRunning")
	// addStateManager("TestEnginePUTRunning", "../../test/resource/engine-run-running.yaml")
	extensionPath, err := global.CopyToTemp("TestEnginePUTRunning", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	addStateManager("TestEnginePUTRunning")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("PUT", "/cr/v1/engine?action=start&extension-name=TestEnginePUTRunning", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleEngine)

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
	handler = http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
	global.RemoveTemp("TestEnginePUTRunning")
}

func TestEngineResetRunning(t *testing.T) {
	t.Log("Entering................. TestEngineResetRunning")
	// addStateManager("TestEngineResetRunning", "../../test/resource/states-reset.yaml")
	extensionPath, err := global.CopyToTemp("TestEngineResetRunning", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	addStateManager("TestEngineResetRunning")

	rr := httptest.NewRecorder()

	req, errState := http.NewRequest("PUT", "/cr/v1/state/task1?status=RUNNING&extension-name=TestEngineResetRunning", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus := http.HandlerFunc(HandleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not set task1 as RUNNING")
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name=TestEngineResetRunning", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Log(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
	global.RemoveTemp("TestEngineResetRunning")
}

func TestEngineReset(t *testing.T) {
	t.Log("Entering................. TestEngineReset")
	// addStateManager("TestEngineReset", "../../test/resource/states-reset.yaml")
	extensionPath, err := global.CopyToTemp("TestEngineReset", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	addStateManager("TestEngineReset")

	rr := httptest.NewRecorder()

	req, errState := http.NewRequest("PUT", "/cr/v1/state/task1?status=FAILED&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus := http.HandlerFunc(HandleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, errState = http.NewRequest("PUT", "/cr/v1/state/task2?status=SKIP&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus = http.HandlerFunc(HandleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	req, errState = http.NewRequest("PUT", "/cr/v1/state/task3?status=READY&extension-name=TestEngineReset", nil)
	if errState != nil {
		t.Fatal(errState)
	}
	handlerStatus = http.HandlerFunc(HandleState)
	handlerStatus.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Can not revert test")
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err = http.NewRequest("PUT", "/cr/v1/engine?action=reset&extension-name=TestEngineReset", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(HandleEngine)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Log(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	global.RemoveTemp("TestEngineReset")
}
