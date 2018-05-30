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
	"strconv"
	"time"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/stateManager"
)

func (cmc *ConfigManagerClient) getLogsByChars(extensionName string, stateName string, firstChar int64, nbChar int64) (string, error) {
	//build the url
	url := "state/" + stateName + "/log?first-char=" + strconv.FormatInt(firstChar, 10)
	url += ";amp&length=" + strconv.FormatInt(nbChar, 10)
	if extensionName != "" {
		url += ";amp&extension-name=" + extensionName
	}
	//Call the rest API
	data, _, err := cmc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	return data, err
}

func (cmc *ConfigManagerClient) getLogs(extensionName string, position int64, stateName string) (int64, error) {
	currentPostion := position
	nbChar := int64(10 * 1024)
	for {
		data, err := cmc.getLogsByChars(extensionName, stateName, currentPostion, nbChar)
		if err != nil {
			return currentPostion, err
		}
		fmt.Print(data)
		currentPostion += int64(len(data))
		if int64(len(data)) != nbChar {
			break
		}
	}
	return currentPostion, nil
}

//Display ongoing logs
// if quiet = true then log are not displayed but the method will wait until the deploy is done.
func (cmc *ConfigManagerClient) follow(extensionName string, position int64, stateName string, quiet bool) error {
	currentPostion := position
	maxRetry := 12
	var errLog error
	//Loop until
	for {
		if maxRetry < 0 {
			errLog = errors.New("Unable to retrieve logs more than 1 min")
		}
		time.Sleep(5 * time.Second)
		var newPos int64
		var err error
		maxRetry = maxRetry - 1
		if !quiet {
			newPos, err = cmc.getLogs(extensionName, currentPostion, stateName)
			if err != nil {
				fmt.Println("\nFailed to get logs (trying again):" + err.Error())
				continue
			}
		}
		if newPos == currentPostion {
			fmt.Print(".")
		}
		if stateName == "cr" {
			currentPostion = newPos
			continue
		}
		//Get last status
		status, err := cmc.getStateStatus(extensionName, stateName)
		if err != nil {
			fmt.Println("\nFailed to get Status (trying again):" + err.Error())
			continue
		}
		maxRetry = 10
		currentPostion = newPos
		if status == stateManager.StateSUCCEEDED {
			break
		}
		if status == stateManager.StateFAILED {
			errLog = errors.New("\nDeployment of " + extensionName + " Failed, state: " + stateName)
			break
		}
	}
	//left over
	time.Sleep(5 * time.Second)
	if !quiet {
		_, err := cmc.getLogs(extensionName, currentPostion, stateName)
		if err != nil {
			fmt.Println("\nFailed to get logs left over:" + err.Error())
		}
	}
	return errLog
}

//GetLogs returns the logs of a given state, if state not provided will retrieve logs from the first state.
// if follow = true the method will loop for new data in the log
// if quiet = true then log are not displayed but the method will wait until the deploy is done.
func (cmc *ConfigManagerClient) GetLogs(extensionName string, stateName string, follow bool, quiet bool) error {
	if stateName == "cr" {
		_, err := cmc.getLogs(extensionName, 0, stateName)
		return err
	}
	//Retrieve list of states and unmarshal
	data, err := cmc.getRestStates(extensionName, "")
	if err != nil {
		return err
	}
	var states stateManager.States
	errUnMarshal := json.Unmarshal([]byte(data), &states)
	if errUnMarshal != nil {
		return errUnMarshal
	}
	//Search the start and end index state to process
	var startStateIndex int
	var endStateIndex int
	if stateName != "" {
		for index, state := range states.StateArray {
			if state.Name == stateName {
				startStateIndex = index
				endStateIndex = index + 1
				break
			}
		}
	} else {
		startStateIndex = 0
		endStateIndex = len(states.StateArray)
	}
	var pos int64
	var currentIndex int
	//display the existing logs for all states
	for index, state := range states.StateArray[startStateIndex:endStateIndex] {
		//Get last status
		status, err := cmc.getStateStatus(extensionName, state.Name)
		if err != nil {
			return err
		}
		if status == stateManager.StateREADY {
			//Sleep to make sure the state changed
			time.Sleep(5 * time.Second)
			//Get last status
			status, err = cmc.getStateStatus(extensionName, state.Name)
			if err != nil {
				return err
			}
		}
		// display logs for status succeed, running and failed
		if status == stateManager.StateSUCCEEDED ||
			status == stateManager.StateRUNNING ||
			status == stateManager.StateFAILED {
			currentIndex = index
			pos, err = cmc.getLogs(extensionName, 0, state.Name)
			if err != nil {
				return err
			}
		}
		//if running or failed nothing else to display (-f will be managed bellow)
		if status == stateManager.StateRUNNING {
			break
		}
		if status == stateManager.StateFAILED {
			return errors.New("\nDeployment of " + extensionName + " failed, state: " + state.Name)
		}
	}
	currentIndex += startStateIndex
	//if -f set follow up
	if follow {
		// manage the remaning logs
		previousEndTime := time.Now()
		for _, state := range states.StateArray[currentIndex:endStateIndex] {
			//wait to make sure the status is up to date.
			//Get last status
			newStateString, err := cmc.getState(extensionName, state.Name)
			if err != nil {
				return err
			}
			var newState stateManager.State
			jsonErr := json.Unmarshal([]byte(newStateString), &newState)
			// status, err := cmc.getStateStatus(extensionName, state.Name)
			if jsonErr != nil {
				return err
			}
			if newState.Status == stateManager.StateSKIP {
				continue
			}
			if newState.Status == stateManager.StateREADY ||
				newState.Status == stateManager.StateFAILED {
				time.Sleep(5 * time.Second)
				newStateString, err = cmc.getState(extensionName, state.Name)
				//				status, err = cmc.getStateStatus(extensionName, state.Name)
				if err != nil {
					return err
				}
				jsonErr = json.Unmarshal([]byte(newStateString), &newState)
				// status, err := cmc.getStateStatus(extensionName, state.Name)
				if jsonErr != nil {
					return err
				}
			}
			//Display RUNNING and failed status.
			// if status == stateManager.StateRUNNING ||
			// 	status == stateManager.StateFAILED {
			startTime, errTimeComv := time.Parse(time.UnixDate, newState.EndTime)
			if errTimeComv != nil ||
				(newState.Status == stateManager.StateSUCCEEDED && startTime.After(previousEndTime)) ||
				newState.Status == stateManager.StateRUNNING ||
				newState.Status == stateManager.StateFAILED {
				err := cmc.follow(extensionName, pos, state.Name, quiet)
				if err != nil {
					return err
				}
			}
			//If status failed no more things to display
			//			if status == stateManager.StateFAILED {
			//				break
			//			}
			pos = 0
		}
	}
	if !quiet {
		fmt.Println("")
	}
	return nil
}
