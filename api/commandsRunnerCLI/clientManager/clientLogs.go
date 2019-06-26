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
	"strconv"
	"time"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
)

func (crc *CommandsRunnerClient) getLogsByChars(extensionName string, stateName string, firstChar int64, nbChar int64) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	//build the url
	url := "state/" + stateName + "/log?extension-name=" + extensionName + "&first-char=" + strconv.FormatInt(firstChar, 10)
	url += "&length=" + strconv.FormatInt(nbChar, 10)
	//Call the rest API
	data, _, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	return data, err
}

func (crc *CommandsRunnerClient) getLogs(extensionName string, position int64, stateName string) (int64, error) {
	currentPostion := position
	nbChar := int64(10 * 1024)
	for {
		data, err := crc.getLogsByChars(extensionName, stateName, currentPostion, nbChar)
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
func (crc *CommandsRunnerClient) follow(extensionName string, position int64, stateName string, quiet bool) error {
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
			newPos, err = crc.getLogs(extensionName, currentPostion, stateName)
			if err != nil {
				fmt.Println("\nFailed to get logs (trying again):" + err.Error())
				continue
			}
		}
		if newPos == currentPostion {
			fmt.Print(".")
		}
		if extensionName == "cr" {
			currentPostion = newPos
			continue
		}
		//Get last status
		status, err := crc.getStateStatus(extensionName, stateName)
		if err != nil {
			fmt.Println("\nFailed to get Status (trying again):" + err.Error())
			continue
		}
		maxRetry = 10
		currentPostion = newPos
		if status == state.StateSUCCEEDED {
			break
		}
		if status == state.StateFAILED {
			errLog = errors.New("\nDeployment of " + extensionName + " Failed, state: " + stateName)
			break
		}
	}
	//left over
	time.Sleep(5 * time.Second)
	if !quiet {
		_, err := crc.getLogs(extensionName, currentPostion, stateName)
		if err != nil {
			fmt.Println("\nFailed to get logs left over:" + err.Error())
		}
	}
	return errLog
}

func isStatePartOfTheCurrentRun(currentState state.State, states state.States) bool {
	return currentState.ExecutedByExtensionName != "" &&
		currentState.ExecutedByExtensionName == states.ExecutedByExtensionName &&
		currentState.ExecutionID != 0 &&
		currentState.ExecutionID == states.ExecutionID
}

func (crc *CommandsRunnerClient) searchNextStatePartOfTheCurrentRun(extensionName string, states state.States, startStateIndex int, endStateIndex int) (*state.State, int, error) {
	var nextState *state.State
	statesAux := states
	nextStateIndex := -1
	// fmt.Println("startStateIndex:" + strconv.Itoa(startStateIndex))
	for {
		for index, stateAux := range statesAux.StateArray[startStateIndex:endStateIndex] {
			// fmt.Println("Index:" + strconv.Itoa(index))
			if isStatePartOfTheCurrentRun(stateAux, statesAux) {
				nextState = &stateAux
				nextStateIndex = index + startStateIndex
				// fmt.Println("nextStateIndex:" + strconv.Itoa(nextStateIndex))
				break
			}
		}
		if nextState != nil {
			break
		}
		// fmt.Println("states.EndTime:" + statesAux.EndTime)
		if statesAux.EndTime != "" {
			return nil, -1, nil
		}
		time.Sleep(5 * time.Second)
		//Retrieve list of states and unmarshal
		data, err := crc.getRestStates(extensionName, "", false, false)
		if err != nil {
			return nil, -1, err
		}
		errUnMarshal := json.Unmarshal([]byte(data), &statesAux)
		if errUnMarshal != nil {
			return nil, -1, errUnMarshal
		}
	}
	return nextState, nextStateIndex, nil
}

//GetLogs returns the logs of a given state, if state not provided will retrieve logs from the first state.
// if follow = true the method will loop for new data in the log
// if quiet = true then log are not displayed but the method will wait until the deploy is done.
func (crc *CommandsRunnerClient) GetLogs(extensionName string, stateName string, follow bool, quiet bool) error {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	if extensionName == "cr" {
		_, err := crc.getLogs(extensionName, 0, stateName)
		return err
	}
	//Retrieve list of states and unmarshal
	data, err := crc.getRestStates(extensionName, "", false, false)
	if err != nil {
		return err
	}
	var states state.States
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
	//Wait that the first states in the range ran or is running
	if follow {
		for {
			statesStarted := false
			for _, stateAux := range states.StateArray[startStateIndex:endStateIndex] {
				if !isStatePartOfTheCurrentRun(stateAux, states) {
					continue
				}
				statesStarted = true
				break
			}
			if statesStarted {
				break
			}
			time.Sleep(5 * time.Second)
			//Retrieve list of states and unmarshal
			data, err := crc.getRestStates(extensionName, "", false, false)
			if err != nil {
				return err
			}
			errUnMarshal := json.Unmarshal([]byte(data), &states)
			if errUnMarshal != nil {
				return errUnMarshal
			}
		}
	}
	var pos int64
	var currentIndex int
	//display the existing logs for all states
	for index, currentState := range states.StateArray[startStateIndex:endStateIndex] {
		//Check if this state is not part of the current execution.
		if !isStatePartOfTheCurrentRun(currentState, states) {
			continue
		}
		//Get last status
		status, err := crc.getStateStatus(extensionName, currentState.Name)
		if err != nil {
			return err
		}
		// display logs for status succeed, running and failed
		if status == state.StateSUCCEEDED ||
			status == state.StateRUNNING ||
			status == state.StateFAILED {
			currentIndex = index
			pos, err = crc.getLogs(extensionName, 0, currentState.Name)
			if err != nil {
				return err
			}
		}
		//if running or failed nothing else to display (-f will be managed bellow)
		if status == state.StateRUNNING {
			break
		}
		if status == state.StateFAILED {
			return errors.New("\nDeployment of " + extensionName + " failed, state: " + currentState.Name)
		}
	}
	currentIndex += startStateIndex
	//if -f set follow up
	if follow {
		// manage the remaning logs
		//		previousEndTime := time.Now()
		for {
			var newState *state.State
			// fmt.Println("Before currentIndex:" + strconv.Itoa(currentIndex))
			newState, currentIndex, err = crc.searchNextStatePartOfTheCurrentRun(extensionName, states, currentIndex, endStateIndex)
			// fmt.Println("After currentIndex:" + strconv.Itoa(currentIndex))
			if err != nil {
				return err
			}
			if currentIndex == -1 {
				break
			}
			if newState.Status == state.StateSUCCEEDED ||
				newState.Status == state.StateRUNNING ||
				newState.Status == state.StateFAILED {
				err := crc.follow(extensionName, pos, newState.Name, quiet)
				if err != nil {
					return err
				}
			}
			currentIndex = currentIndex + 1
			pos = 0
		}
	}
	if !quiet {
		fmt.Println("")
	}
	return nil
}
