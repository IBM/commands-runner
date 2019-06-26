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
package clientManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/status"
)

//GetCMStatus returns the configManager status
func (crc *CommandsRunnerClient) GetCMStatus() (string, error) {
	url := "status"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get status: " + data + ", please check log for more information")
	}
	//Convert the config to object
	var statuses status.Statuses
	jsonErr := json.Unmarshal([]byte(data), &statuses)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		return "", jsonErr
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		out := ""
		for _, status := range statuses {
			out += fmt.Sprintf("%s: %s\n", status.Name, status.Status)
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//SetCMStatus set a status
func (crc *CommandsRunnerClient) SetCMStatus(name string, status string) (string, error) {
	url := "status?name=" + url.QueryEscape(name) + "&status=" + url.QueryEscape(status)
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to set status: " + data + ", please check log for more information")
	}
	//Convert the config to object
	return crc.GetCMStatus()
}
