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

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
)

func (crc *CommandsRunnerClient) getState(extensionName string, stateName string) (string, error) {
	//build url
	url := "state/" + stateName
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest api
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get states: " + data + ", please check log for more information")
	}
	return data, err
}

func (crc *CommandsRunnerClient) getStateStatus(extensionName string, stateName string) (string, error) {
	currentGetState, errGetStatus := crc.getState(extensionName, stateName)
	if errGetStatus != nil {
		return "", errGetStatus
	}
	var currentState state.State
	jsonErr := json.Unmarshal([]byte(currentGetState), &currentState)
	if jsonErr != nil {
		return "", jsonErr
	}
	return currentState.Status, nil
}

//GetState returns the status of a given state
func (crc *CommandsRunnerClient) GetState(extensionName string, stateName string) (string, error) {
	//Check is state name is provided
	if stateName == "" {
		err := errors.New("--state|-s is required")
		return "", err
	}
	data, err := crc.getState(extensionName, stateName)
	if err != nil {
		return "", err
	}
	//Convert in text format otherwize return the json
	if crc.OutputFormat == "text" {
		var state state.State
		jsonErr := json.Unmarshal([]byte(data), &state)
		if jsonErr != nil {
			return "", jsonErr
		}
		out := fmt.Sprintf("State      : %s\n", state.Label)
		out += fmt.Sprintf("State name : %s\n", state.Name)
		out += fmt.Sprintf("State sttus: %s\n", state.Status)
		out += fmt.Sprintf("Start time : %s\n", state.StartTime)
		out += fmt.Sprintf("End time   : %s\n", state.EndTime)
		out += fmt.Sprintf("Reason     : %s\n", state.Reason)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//SetState
func (crc *CommandsRunnerClient) SetState(extensionName string, stateName string, newStatus string, scriptTimeout string) error {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	//Check is state name is provided
	if stateName == "" {
		err := errors.New("--state|-s is required")
		return err
	}
	query := ""
	if newStatus != "" {
		query = "status=" + newStatus
	}
	if scriptTimeout != "" {
		if query != "" {
			query = query + "&"
		}
		query = query + "script-timeout=" + scriptTimeout
	}
	//build url
	url := "state/" + stateName + "?" + query
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	//Call the rest api
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return err
	}
	if errCode != http.StatusOK {
		return errors.New(data)
	}
	return nil
}
