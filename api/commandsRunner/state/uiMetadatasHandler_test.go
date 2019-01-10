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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
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
