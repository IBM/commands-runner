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
	"errors"
	"net/http"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//ResetEngine resets states, all not "SKIP" states will be set to "READY".
//No running state must exit
func (crc *CommandsRunnerClient) ResetEngine(extensionName string) (string, error) {
	//Build url
	url := "engine?action=reset"
	if extensionName != "" {
		url += ";amp&extension-name=" + extensionName
	}
	//Call rest api
	_, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to reset engine, please check logs")
	}
	return "", nil
}

//IsRunningEngine checks if engine is running".
//No running state must exit
func (crc *CommandsRunnerClient) IsRunningEngine(extensionName string) (string, error) {
	//Build url
	url := "engine"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call rest api
	_, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK && errCode != http.StatusProcessing {
		return "", errors.New("Unable to retrieve the status")
	}
	if errCode == http.StatusOK {
		return "State engine is not running\n", nil
	}
	return "State engine is running\n", nil
}

//StartEngine returns the states
func (crc *CommandsRunnerClient) StartEngine(extensionName string, fromState string, toState string) (string, error) {
	//build url
	url := "engine?action=start"
	if extensionName != "" {
		url += ";amp&extension-name=" + extensionName
	}
	if fromState != "" {
		url += ";amp&from-state=" + fromState
	}
	if toState != "" {
		url += ";amp&to-state=" + toState
	}
	//Call rest api
	_, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		if errCode == http.StatusConflict {
			return "", errors.New("Engine already running")
		}
		return "", errors.New("Unable to start the engine, please check the logs")
	}
	return "", nil
}
