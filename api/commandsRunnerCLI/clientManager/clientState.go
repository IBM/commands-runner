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
package clientManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
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
		return "", errors.New("Unable to get states, please check logs")
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
			query = query + "&amp;"
		}
		query = query + "script-timeout=" + scriptTimeout
	}
	//build url
	url := "state/" + stateName + "?" + query
	if extensionName != "" {
		url += "&amp;extension-name=" + extensionName
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
