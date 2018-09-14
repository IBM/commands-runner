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
package uiMetadata

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUIConfigEndpointFailed(t *testing.T) {
	t.Log("Entering................. TestGetUIConfigEndpointFailed")
	req, err := http.NewRequest("GET", "/cr/v1/uimetadata/does-not-exist/hello", nil)
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
