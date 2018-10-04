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
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetUIConfigEndpointFailed(t *testing.T) {
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/resource/extensions/")
	req, err := http.NewRequest("GET", "/cr/v1/uimetadata?extension-name=does-not-exist&ui-metadata-name=hello", nil)
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
}

func TestGetUIConfigEndpointSuccess(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/resource/extensions/")
	req, err := http.NewRequest("GET", "/cr/v1/uimetadata?extension-name=ext-template&ui-metadata-name=test-ui", nil)
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
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	t.Log(rr.Body)
}
