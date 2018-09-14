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
package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
)

//Commenting because there is a concurrence issue when we use 2 different states file
func TestSaveConfig(t *testing.T) {
	t.Log("Entering................. TestSaveConfig")
	SetConfigPath("../../test/resource/ConfigDir")
	state.SetExtensionPath("../../test/resource/extensions")
	body, err := os.Open("../../test/resource/config-test-save.yml")
	defer body.Close()
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/cr/v1/config?extension-name=config-handler-test", body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Check Body")
	//data, _ := ioutil.ReadAll(req.Body)
	//log.Debug(string(data))
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleConfig)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}

func TestGetConfig(t *testing.T) {
	t.Log("Entering................. TestSaveConfig")
	//log.SetLevel(log.DebugLevel)
	SetConfigPath("../../test/resource")
	state.SetExtensionPath("../../test/resource/extensions")
	bckConfigFileName := global.ConfigYamlFileName
	SetConfigFileName("config-test-save.yml")

	req, err := http.NewRequest("GET", "/cr/v1/config?extension-name=config-handler-test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleConfig)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
	SetConfigFileName(bckConfigFileName)
}

func TestGetConfigCustomized(t *testing.T) {
	t.Log("Entering................. TestGetConfigCustomized")
	//log.SetLevel(log.DebugLevel)
	SetConfigPath("../../test/resource")
	state.SetExtensionPath("../../test/resource/extensions")
	bckConfigFileName := global.ConfigYamlFileName
	SetConfigFileName("uiconfig-test-save.yml")
	bckConfigRootKey := global.ConfigRootKey
	SetConfigRootKey("uiconfig")

	req, err := http.NewRequest("GET", "/cr/v1/config?extension-name=config-handler-test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleConfig)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
	SetConfigFileName(bckConfigFileName)
	SetConfigRootKey(bckConfigRootKey)
}
