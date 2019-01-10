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
package commandsRunner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func TestGetSettings(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	global.DefaultExtensionName = "test-extension"
	global.ConfigRootKey = "myconfig"

	req, err := http.NewRequest("GET", "/cr/v1/cr/settings", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleCR)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}

	var settings Settings
	body, err := ioutil.ReadAll(rr.Body)
	err = json.Unmarshal(body, &settings)
	if err != nil {
		t.Error(err)
	}
	t.Log(settings)
	if settings.ConfigRootKey != "myconfig" || settings.DefaultExtensionName != "test-extension" {
		t.Errorf("Not expected response: %v", settings)
	}
}
