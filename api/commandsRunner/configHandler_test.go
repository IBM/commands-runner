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
	"os"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/configManager"
)

//Commenting because there is a concurrence issue when we use 2 different states file
func TestSaveConfig(t *testing.T) {
	t.Log("Entering................. TestSaveConfig")
	configManager.SetConfigPath("../test/resource/CloudFoundry")
	body, err := os.Open("../test/resource/uiconfig-test-save.yml")
	defer body.Close()
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/cr/v1/bmxconfig", body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Check Body")
	//data, _ := ioutil.ReadAll(req.Body)
	//log.Debug(string(data))
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleConfig)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}
