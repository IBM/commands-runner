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

	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func TestGetUIMetadatasConfigEndpointFailed(t *testing.T) {
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIMetadatasConfigEndpointFailed", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/uimetadatas?extension-name=does-not-exist&ui-metadata-name=hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUIMetadata)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	global.RemoveTemp("TestGetUIMetadatasConfigEndpointFailed")
}

func TestGetUIMetadatasConfigEndpointSuccess(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIMetadatasConfigEndpointSuccess", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/uimetadatas?extension-name=ext-template&ui-metadata-name=test-ui", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUIMetadatas)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	t.Log(rr.Body)
	global.RemoveTemp("TestGetUIMetadatasConfigEndpointSuccess")
}

func TestGetUIMetadatasAllConfigEndpointSuccess(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	SetExtensionsEmbeddedFile("../../test/data/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIMetadatasAllConfigEndpointSuccess", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	req, err := http.NewRequest("GET", "/cr/v1/uimetadatas?names-only=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUIMetadatas)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	t.Log(rr.Body)
	global.RemoveTemp("TestGetUIMetadatasAllConfigEndpointSuccess")
}
