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
package configManagerClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/stateManager"
)

func (cmc *ConfigManagerClient) getState(extensionName string, stateName string) (string, error) {
	//build url
	url := "state/" + stateName
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest api
	data, errCode, err := cmc.restCall(http.MethodGet, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get states, please check logs")
	}
	return data, err
}

func (cmc *ConfigManagerClient) getStateStatus(extensionName string, stateName string) (string, error) {
	currentGetState, errGetStatus := cmc.getState(extensionName, stateName)
	if errGetStatus != nil {
		return "", errGetStatus
	}
	var currentState stateManager.State
	jsonErr := json.Unmarshal([]byte(currentGetState), &currentState)
	if jsonErr != nil {
		return "", jsonErr
	}
	return currentState.Status, nil
}

//GetState returns the status of a given state
func (cmc *ConfigManagerClient) GetState(extensionName string, stateName string) (string, error) {
	//Check is state name is provided
	if stateName == "" {
		err := errors.New("--state|-s is required")
		return "", err
	}
	data, err := cmc.getState(extensionName, stateName)
	if err != nil {
		return "", err
	}
	//Convert in text format otherwize return the json
	if cmc.OutputFormat == "text" {
		var state stateManager.State
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
	return cmc.convertJSONOrYAML(data)
}

//SetState
func (cmc *ConfigManagerClient) SetState(extensionName string, stateName string, newStatus string, scriptTimeout string) error {
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
			query = query + ";amp&"
		}
		query = query + "script-timeout=" + scriptTimeout
	}
	//build url
	url := "state/" + stateName + "?" + query
	if extensionName != "" {
		url += ";amp&extension-name=" + extensionName
	}
	//Call the rest api
	data, errCode, err := cmc.restCall(http.MethodPut, url, nil, nil)
	if err != nil {
		return err
	}
	if errCode != http.StatusOK {
		return errors.New(data)
	}
	return nil
}
