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

func (crc *CommandsRunnerClient) getRestStates(extensionName string, status string, extensionOnly bool, recursive bool) (string, error) {
	//build url
	url := "states"
	url += "?extensions-only=" + strconv.FormatBool(extensionOnly)
	url += "&recursive=" + strconv.FormatBool(recursive)
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	if status != "" {
		url += "&status=" + status
	}
	//Call rest api
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get states: " + data + ", please check log for more information")
	}
	return data, err
}

//GetStates returns the states having a specific status
func (crc *CommandsRunnerClient) GetStates(extensionName string, status string, extensionOnly bool, recursive bool) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	data, err := crc.getRestStates(extensionName, status, extensionOnly, recursive)
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
			out += fmt.Sprintf("Deleted   : %t\n", state.Deleted)
			out += fmt.Sprintf("Next run  : %t\n", state.NextRun)
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
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
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
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, file, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to save the states file: " + data + ", please check log for more information")
	}
	return "", nil
}

//Set a state file
func (crc *CommandsRunnerClient) SetStatesStatuses(extensionName string, newStatus string, fromState string, fromInclude bool, toState string, toInclude bool) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
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
		return "", errors.New("Unable to set States statuses: " + data + ", please check log for more information")
	}
	data, err = crc.GetStates(extensionName, "", false, false)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (crc *CommandsRunnerClient) InsertStateStates(extensionName string, pos int, stateName string, before bool, statePath string, insertExtensionName string, overwrite bool) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	if statePath == "" && insertExtensionName == "" {
		err := errors.New("A state file or extension name missing")
		return "", err
	}
	url := "states?action=insert&pos=" + strconv.Itoa(pos) + "&before=" + strconv.FormatBool(before) + "&overwrite=" + strconv.FormatBool(overwrite)
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	if stateName != "" {
		url += "&state-name=" + stateName
	}
	if insertExtensionName != "" {
		url += "&insert-extension-name=" + insertExtensionName
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
	data, err = crc.GetStates(extensionName, "", false, false)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (crc *CommandsRunnerClient) DeleteStateStates(extensionName string, pos int, stateName string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	url := "states?action=delete&pos=" + strconv.Itoa(pos)
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	if stateName != "" {
		url += "&state-name=" + stateName
	}
	//Call the rest API
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to delete state at position: " + strconv.Itoa(pos) + " in states file\n" + data)
	}
	data, err = crc.GetStates(extensionName, "", false, false)
	if err != nil {
		return "", err
	}
	return data, nil
}
