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
package status

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetStatusesEndpoint(t *testing.T) {
	//log.SetLevel(log.InfoLevel)
	t.Log("Entering................. TestGetStatusesEndpoint")
	req, err := http.NewRequest("GET", "/cr/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStatus)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	// Check the response body is what we expect.
	expected := "{\"cr_status\":{\"name\":\"cr_status\",\"value\":\"Initialization\"},\"log_level\":{\"name\":\"log_level\",\"value\":\"info\"}}"
	got := strings.TrimRight(rr.Body.String(), "\n")
	t.Log(string(expected))
	t.Log(got)
	if got != string(expected) {
		t.Errorf("handler returned unexpected body: got %v- want %v-",
			rr.Body.String(), string(expected))
	}

}

func TestSetStatusesEndpoint(t *testing.T) {
	//log.SetLevel(log.InfoLevel)
	t.Log("Entering................. TestSetStatusesEndpoint")
	req, err := http.NewRequest("PUT", "/cr/v1/status?name="+CMStatus+"&status=newStatus", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStatus)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	req, err = http.NewRequest("GET", "/cr/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, %v want %v",
			status, rr.Body, http.StatusOK)
	}
	// Check the response body is what we expect.
	expected := "{\"cr_status\":{\"name\":\"" + CMStatus + "\",\"value\":\"newStatus\"},\"log_level\":{\"name\":\"log_level\",\"value\":\"info\"}}"
	got := strings.TrimRight(rr.Body.String(), "\n")
	t.Log(string(expected))
	t.Log(got)
	if got != string(expected) {
		t.Errorf("handler returned unexpected body: got %v- want %v-",
			rr.Body.String(), string(expected))
	}

}

func TestGetStatusesEndpointWrongMethod(t *testing.T) {
	t.Log("Entering................. TestGetStatusesEndpointWrongMethod")
	req, err := http.NewRequest("POST", "/cr/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleStatus)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Unsupported method... an error is raised")
	}

}
