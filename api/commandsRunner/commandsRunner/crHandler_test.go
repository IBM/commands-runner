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
package commandsRunner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
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
