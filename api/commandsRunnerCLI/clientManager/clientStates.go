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
	"io"
	"net/http"
	"os"
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
)

func (crc *CommandsRunnerClient) getRestStates(extensionName string, status string) (string, error) {
	//build url
	url := "states"
	if status != "" {
		url += "?status=" + status
		if extensionName != "" {
			url += "&amp;extension-name=" + extensionName
		}
	} else {
		if extensionName != "" {
			url += "?extension-name=" + extensionName
		}
	}
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get states, please check logs")
	}
	return data, err
}

//GetStates returns the states having a specific status
func (crc *CommandsRunnerClient) GetStates(extensionName string, status string) (string, error) {
	data, err := crc.getRestStates(extensionName, status)
	if err != nil {
		return "", err
	}

	//Convert to text otherwize return the json
	if crc.OutputFormat == "text" {
		var states state.States
		jsonErr := json.Unmarshal([]byte(data), &states)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		out := ""
		for _, state := range states.StateArray {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("State name: %s\n", state.Name)
			out += fmt.Sprintf("Label     : %s\n", state.Label)
			out += fmt.Sprintf("Phase     : %s\n", state.Phase)
			out += fmt.Sprintf("Script    : %s\n", state.Script)
			out += fmt.Sprintf("Timeout   : %d\n", state.ScriptTimeout)
			out += fmt.Sprintf("LogPath   : %s\n", state.LogPath)
			out += fmt.Sprintf("Status    : %s\n", state.Status)
			out += fmt.Sprintf("Start time: %s\n", state.StartTime)
			out += fmt.Sprintf("End time  : %s\n", state.EndTime)
			out += fmt.Sprintf("Reason    : %s\n", state.Reason)
			out += fmt.Sprintf("Protected : %t\n", state.Protected)
			out += fmt.Sprintf("Deleted : %t\n", state.Deleted)
			for _, stateToRun := range state.StatesToRerun {
				out += fmt.Sprintf("- State to rerun : %s\n", stateToRun)
			}
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//Set a state file
func (crc *CommandsRunnerClient) SetStates(extensionName string, statesPath string, overwrite bool) (string, error) {
	if statesPath == "" {
		err := errors.New("states file missing")
		return "", err
	}
	url := "states?overwrite=" + strconv.FormatBool(overwrite)
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	var file io.Reader
	fileOS, errFile := os.Open(statesPath)
	if errFile != nil {
		return "", errFile
	}
	file = fileOS
	//Call the rest API
	_, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, file, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to save the states file, please check the format")
	}
	return "", nil
}

//Set a state file
func (crc *CommandsRunnerClient) SetStatesStatuses(extensionName string, newStatus string, fromState string, fromInclude bool, toState string, toInclude bool) (string, error) {
	if newStatus == "" {
		err := errors.New("new status is missing")
		return "", err
	}
	url := "states?action=set-statuses&status=" + newStatus
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	if fromState != "" {
		url += "&from-state-name=" + fromState + "&from-include=" + strconv.FormatBool(fromInclude)
	}
	if toState != "" {
		url += "&to-state-name=" + toState + "&to-include=" + strconv.FormatBool(toInclude)
	}
	//Call the rest API
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to set States statuses:" + data)
	}
	data, err = crc.GetStates(extensionName, "")
	if err != nil {
		return "", err
	}
	return data, nil
}

func (crc *CommandsRunnerClient) InsertStateStates(extensionName string, pos int, stateName string, before bool, statePath string, insertExtensionName string) (string, error) {
	if statePath == "" && insertExtensionName == "" {
		err := errors.New("A state file or extension name missing")
		return "", err
	}
	url := "states?action=insert&amp;pos=" + strconv.Itoa(pos) + "&amp;before=" + strconv.FormatBool(before)
	if extensionName != "" {
		url += "&amp;extension-name=" + extensionName
	}
	if stateName != "" {
		url += "&amp;state-name=" + stateName
	}
	if insertExtensionName != "" {
		url += "&amp;insert-extension-name=" + insertExtensionName
	}
	//Call the rest API
	var file io.Reader
	if statePath != "" {
		fileOS, errFile := os.Open(statePath)
		if errFile != nil {
			return "", errFile
		}
		file = fileOS
	} else {
		file = nil
	}
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, file, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to insert state in states file\n" + data)
	}
	data, err = crc.GetStates(extensionName, "")
	if err != nil {
		return "", err
	}
	return data, nil
}

func (crc *CommandsRunnerClient) DeleteStateStates(extensionName string, pos int, stateName string) (string, error) {
	url := "states?action=delete&amp;pos=" + strconv.Itoa(pos)
	if extensionName != "" {
		url += "&amp;extension-name=" + extensionName
	}
	if stateName != "" {
		url += "&amp;state-name=" + stateName
	}
	//Call the rest API
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to delete state at position" + strconv.Itoa(pos) + " in states file\n" + data)
	}
	data, err = crc.GetStates(extensionName, "")
	if err != nil {
		return "", err
	}
	return data, nil
}
