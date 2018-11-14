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
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//ResetEngine resets states, all not "SKIP" states will be set to "READY".
//No running state must exit
func (crc *CommandsRunnerClient) ResetEngine(extensionName string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	//Build url
	url := "engine?action=reset"
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to reset engine: " + data + ", please check log for more information")
	}
	return "", nil
}

//IsRunningEngine checks if engine is running".
//No running state must exit
func (crc *CommandsRunnerClient) IsRunningEngine(extensionName string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	//Build url
	url := "engine"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK && errCode != http.StatusCreated {
		return "", errors.New("Unable to retrieve the status:" + data)
	}
	if errCode == http.StatusOK {
		return "State engine is not running\n", nil
	}
	return "State engine is running\n", nil
}

//StartEngine returns the states
func (crc *CommandsRunnerClient) StartEngine(extensionName string, fromState string, toState string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	//build url
	url := "engine?action=start"
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	if fromState != "" {
		url += "&from-state=" + fromState
	}
	if toState != "" {
		url += "&to-state=" + toState
	}
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		if errCode == http.StatusConflict {
			return "", errors.New("Engine already running: " + data + ", please check log for more information")
		}
		return "", errors.New("Unable to start the engine: " + data + ", please check the logs")
	}
	return "", nil
}

//MockEngine set/unset engine mock mode.
//No running state must exit
func (crc *CommandsRunnerClient) SetMockEngine(mock bool) (string, error) {
	//Build url
	url := "engine?action=mock&mock=" + strconv.FormatBool(mock)
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to set the mock mode for the engine: " + data + ", please check log for more information")
	}
	return "", nil
}

func (crc *CommandsRunnerClient) GetMockEngine() (string, error) {
	//Build url
	url := "engine?action=mock"
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get the mock mode for the engine: " + data + ", please check log for more information")
	}
	if crc.OutputFormat == "text" {
		var mock state.Mock
		jsonErr := json.Unmarshal([]byte(data), &mock)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		out := fmt.Sprintf("mock : %t\n", mock.Mock)
		return out, nil
	}

	return crc.convertJSONOrYAML(data)
}
