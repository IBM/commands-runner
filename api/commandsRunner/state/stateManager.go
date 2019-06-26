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
package state

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/commands-runner/api/commandsRunner/logger"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"

	"github.com/IBM/commands-runner/api/commandsRunner/global"

	"gonum.org/v1/gonum/graph"

	"github.com/olebedev/config"
	log "github.com/sirupsen/logrus"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	"github.com/go-yaml/yaml"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

const FirstState = "FirstState"
const LastState = "LastState"
const StateREADY = "READY"
const StateFAILED = "FAILED"
const StateSUCCEEDED = "SUCCEEDED"
const StateRUNNING = "RUNNING"
const StateSKIP = "SKIP"
const StatePREPROCESSING = "PREPROCESSING"

const StatesFileErrorMessagePattern = "STATES_FILE_ERROR_MESSAGE:"

const PhaseAtEachRun = "AtEachRun"

type Mock struct {
	Mock bool `yaml:"mock" json:"mock"`
}

type State struct {
	//Name name of the state
	Name string `yaml:"name" json:"name"`
	//Phase If the value is set to "AtEachRun" then the state will be executed each time the commands-runner is launched and this independently of its status except if status is "SKIP"
	Phase string `yaml:"phase" json:"phase"`
	//Label A more human readable name for the state (default: name)
	Label string `yaml:"label" json:"label"`
	//LogPath Location of the file that will collect the stdin/stderr. (default: extensionPath+extenstionName.log)
	LogPath string `yaml:"log_path" json:"log_path"`
	//Status The current state status (default: READY)
	Status string `yaml:"status" json:"status"`
	//StartTime If not empty, it contains the last time when the state was executed
	StartTime string `yaml:"start_time" json:"start_time"`
	//EndTime if not empty, it contains the last end execution time of the state
	EndTime string `yaml:"end_time" json:"end_time"`
	//Reason if not empty, it contains the reason of execution failure
	Reason string `yaml:"reason" json:"reason"`
	//Script The command or script to execute. The path must be an absolute path
	Script string `yaml:"script" json:"script"`
	//ScriptTimeout The maximum duration of the state execution, after that duration a timeout error will be produced.
	ScriptTimeout int `yaml:"script_timeout" json:"script_timeout"`
	//Protected If true the commands-runner end-user will be not be able to delete the state.
	Protected bool `yaml:"protected" json:"protected"`
	//Deleted If true the corresponding state will be deleted when merging with an existing states file.
	Deleted bool `yaml:"deleted" json:"deleted"`
	//PrerequisiteStates  if the current states is READY/FAILED then those listed will be set to READY too.
	//Each listed state must be before the current state.
	PrerequisiteStates []string `yaml:"prerequisite_states" json:"prerequisite_states"`
	//StatesToRerun if the current state is READY/FAILED then those listed will be set to READY too.
	//Each listed state must be after the current state.
	StatesToRerun []string `yaml:"states_to_rerun" json:"states_to_rerun"`
	//RerunOnRunOfStates if one state of this list is READY/FAILED then the current state will be set to READY.
	//Each listed state must be before the current state.
	RerunOnRunOfStates []string `yaml:"rerun_on_run_of_states" json:"rerun_on_run_of_states"`
	//CalculatedStatesToRerun (not-persisted/used internally) Calculated list of states to rerun
	CalculatedStatesToRerun []string `yaml:"-" json:"-"`
	//PreviousStates List of previous states, this is not taken into account for the topology sort.
	PreviousStates []string `yaml:"previous_states" json:"previous_states"`
	//NextStates List of next states, this determines the order in which the states will get executed. A topological sort is used to determine the order.
	//If this attribute is not defined in any state of the states file then the NextStates for each state will be set by default to the next state in the states file.
	NextStates []string `yaml:"next_states" json:"next_states"`
	//ExecutedByExtensionName is set at execution time with the value of sm.ExecutedByExtensionName.
	ExecutedByExtensionName string `yaml:"executed_by_extension_name" json:"executed_by_extension_name"`
	//The execution sequence id for that specific launch
	ExecutionID int `yaml:"execution_id" json:"execution_id"`
	//This is true if the state will run based on status, PrerequisiteStates, RerunOnRunOfStates and StatesToRerun
	NextRun bool `yaml:"next_run" json:"next_run"`
	//This is true when the state is a extension
	IsExtension bool `yaml:"is_extension" json:"is_extension"`
}

type States struct {
	StateArray    []State `yaml:"states" json:"states"`
	ExtensionName string  `yaml:"extension_name" json:"extension_name"`
	//Parent extension name, this is set when the extension is inserted into another extension.
	//Empty if not inserted.
	ParentExtensionName string `yaml:"parent_extension_name" json:"parent_extension_name"`
	//ExecutedByExtensionName is set at execution time and contains the name of the extension which launch the current extension.
	//If extension A contains extension B contains extension C then it contains extension A name.
	//It is NOT the direct parent extension name.
	ExecutedByExtensionName string `yaml:"executed_by_extension_name" json:"executed_by_extension_name"`
	//The execution sequence id for that specific launch
	ExecutionID int `yaml:"execution_id" json:"execution_id"`
	//StartTime If not empty, it contains the last time when the state was executed
	StartTime string `yaml:"start_time" json:"start_time"`
	//EndTime if not empty, it contains the last end execution time of the state
	EndTime string `yaml:"end_time" json:"end_time"`
	//Status states status
	Status     string `yaml:"status" json:"status"`
	StatesPath string `yaml:"-" json:"-"`
	mux        *sync.Mutex
}

var crLogTempFile *os.File

var stateManagers map[string]States

//initialize the map of stateManagers
func init() {
	stateManagers = make(map[string]States)
}

//get the statesFile path
func getStatePath(extensionName string) (string, error) {
	if extensionName == "" {
		return "", errors.New("extensionName is empty")
	}
	var statesPathAux string
	statesPathAux = GetRootExtensionPath(GetExtensionsPath(), extensionName)
	statesPathAux = filepath.Join(statesPathAux, global.StatesFileName)
	return statesPathAux, nil
}

//Add a state manager to the map, directly used only for test method.
func addStateManager(extensionName string) {
	log.Debug("Entering in addStateManager")
	log.Debug("Extension name: " + extensionName)
	sm := newStateManager(extensionName)
	stateManagers[extensionName] = *sm
	log.Debug("State Manager added for " + extensionName)
}

//Remove a stateManager
func removeStateManager(extensionName string) error {
	delete(stateManagers, extensionName)
	return nil
}

//Find a stateManager based on the extensionNAme
func getStatesManager(extensionName string) (*States, error) {
	log.Debug("Entering in getStatesManager")
	log.Debug("ExtensionName: " + extensionName)
	if val, ok := stateManagers[extensionName]; ok {
		statePath, _ := getStatePath(extensionName)
		log.Debug("statePath:" + statePath)
		val.StatesPath = statePath
		return &val, nil
	}
	return nil, errors.New("stateManager not found for " + extensionName)
}

//Search for a stateManager and if not found create it
func GetStatesManager(extensionName string) (*States, error) {
	log.Debug("Entering in getAddStateManagersss")
	log.Debug("Search for manager: " + extensionName)
	sm, err := getStatesManager(extensionName)
	if err == nil {
		log.Debug("Manager already exists, returning it")
		return sm, nil
	}
	log.Debug("Manager doesn't exist, creating")
	addStateManager(extensionName)
	log.Debug("returning created manager")
	return getStatesManager(extensionName)
}

//NewClient creates a new client
func newStateManager(extensionName string) *States {
	log.Debug("Entering... NewStateManager")
	//Set the default values
	states := &States{
		StateArray:              make([]State, 0),
		ExtensionName:           extensionName,
		ExecutedByExtensionName: "",
		ExecutionID:             0,
		StartTime:               "",
		EndTime:                 "",
		Status:                  "",
		ParentExtensionName:     "",
		StatesPath:              "",
		mux:                     &sync.Mutex{},
	}
	return states
}

func (sm *States) isCustomStatePath() bool {
	log.Debug("Entering... isCustomStatePath")
	log.Debug("StatePath:" + sm.StatesPath)
	if strings.Contains(sm.StatesPath, "/custom/") {
		return true
	}
	return false
}

func (sm *States) lock() {
	log.Debug("Lock states")
	sm.mux.Lock()
}
func (sm *States) unlock() {
	log.Debug("Unlock states")
	sm.mux.Unlock()
}

// Read state
// We can not use defer on sm.unlock because it is not an argument of the function
func (sm *States) readStates() error {
	log.Debug("Entering... readStates")
	log.Debug("statesPath... " + sm.StatesPath)
	log.Debugf("sm address %p:", &sm)
	statesData, err := ioutil.ReadFile(sm.StatesPath)
	//	log.Debugf("StatesData=%s", statesData)
	if err != nil {
		return err
	}
	log.Debug("states has been read")
	//	log.Debug("States:\n" + string(statesData))
	// Parse state file into the States structure
	err = yaml.Unmarshal(statesData, &sm)
	log.Debugf("sm address %p:", &sm)
	if err != nil {
		return err
	}
	// err = sm.setCalculatedStatesToRerun()
	// if err != nil {
	// 	return err
	// }
	sm.setDefaultValues()
	//	log.Debug("States:\n" + string(statesData))
	log.Debug("Exiting... readStates")
	return nil
}

//Set Calculated statesToRerun
func (sm *States) setCalculatedStatesToRerun() error {
	log.Debug("Entering... setCalculatedStatesToRerun")
	//	isNextStateMigrationDone := false
	for index := range sm.StateArray {
		sm.StateArray[index].CalculatedStatesToRerun = append(sm.StateArray[index].CalculatedStatesToRerun, sm.StateArray[index].StatesToRerun...)
		sm.StateArray[index].CalculatedStatesToRerun = append(sm.StateArray[index].CalculatedStatesToRerun, sm.StateArray[index].PrerequisiteStates...)
		for _, stateName := range sm.StateArray[index].RerunOnRunOfStates {
			state, err := sm._getState(stateName)
			//if state is found do update
			if err == nil {
				state.CalculatedStatesToRerun = append(state.CalculatedStatesToRerun, sm.StateArray[index].Name)
			}
		}
	}
	for index := range sm.StateArray {
		log.Debugf("%s - %v", sm.StateArray[index].Name, sm.StateArray[index].CalculatedStatesToRerun)
	}
	log.Debug("Exiting... setCalculatedStatesToRerun")
	return nil
}

//Set default value for states
func (sm *States) setDefaultValues() {
	log.Debug("Entering... setDefaultValues")
	log.Debug("StatePath:" + sm.StatesPath)
	//	isNextStateMigrationDone := false
	for index := range sm.StateArray {
		//		log.Debug("Check state:" + sm.StateArray[index].Name)
		//		log.Debug("Check Label")
		if sm.StateArray[index].Label == "" {
			sm.StateArray[index].Label = sm.StateArray[index].Name
			//			log.Debug("Set state.Label to " + sm.StateArray[index].Label)
		}
		//		log.Debug("Check status")
		if sm.StateArray[index].Status == "" {
			sm.setStateStatus(sm.StateArray[index], StateREADY, true)
		}
		//		log.Debug("Check LogPath/Script")

		if sm.StateArray[index].LogPath == "" {
			dir := GetExtensionsLogsPathEmbedded()
			if sm.isCustomStatePath() {
				log.Debug("Customer extension")
				dir = GetExtensionsLogsPathCustom()
			} else {
				log.Debug("Embbeded extension")
			}
			log.Debug("ExtensionLogPath:" + dir)
			logDir := filepath.Join(dir, sm.ExtensionName, sm.StateArray[index].Name+".log")
			sm.StateArray[index].LogPath = logDir
			log.Debug("Set state.LogPath to " + sm.StateArray[index].LogPath)
		}
		if sm.StateArray[index].ScriptTimeout == 0 {
			sm.StateArray[index].ScriptTimeout = 60
		}
		// sm.StateArray[index].IsExtension = false
	}
	log.Debug("Exiting... setDefaultValues")
}

//setNextStates sets the next states in case of migration. Migration is detected if all NextStates array are empty.
func (sm *States) setNextStates() {
	for index := range sm.StateArray {
		if !sm.StateArray[index].Deleted && index < len(sm.StateArray)-1 {
			indexNext := sm.searchNextStates(index)
			if indexNext != -1 {
				stateNext := sm.StateArray[indexNext]
				if len(sm.StateArray[index].NextStates) == 0 {
					sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates, stateNext.Name)
				}
			}
		}
	}
}

//copyStateToRerunToNextStates copy the StateToRerun to the NextStates.
//This is done for migration only.
func (sm *States) copyStateToRerunToNextStates() {
	for index := range sm.StateArray {
		if !sm.StateArray[index].Deleted {
			for _, stateRerun := range sm.StateArray[index].StatesToRerun {
				if !sm.isInNextStates(sm.StateArray[index], stateRerun) {
					sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates, stateRerun)
				}
			}
		}
	}
}

//searchNextStates searches the next non-delete states after the current state index.
func (sm *States) searchNextStates(start int) int {
	for i := start + 1; i < len(sm.StateArray); i++ {
		if !sm.StateArray[i].Deleted {
			return i
		}
	}
	return -1
}

//Write states
// We can not use defer on sm.unlock because it is not an argument of the function
func (sm *States) writeStates() error {
	log.Debug("Entering... writeStates")
	log.Debug("Marshal states")
	//	log.Debug(sm )
	//	sm.lock()
	// defer sm.unlock()
	statesData, err := sm.convert2ByteArray()
	log.Debugf("statesData: %s", string(statesData))
	if err != nil {
		return err
	}
	//log.Debugf("Write states to %s", statesPath)
	dir := filepath.Dir(sm.StatesPath)
	errMkDir := os.MkdirAll(dir, 0777)
	if errMkDir != nil {
		return errMkDir
	}
	err = ioutil.WriteFile(sm.StatesPath, statesData, 0666)
	if err != nil {
		return err
	}
	return nil
}

//convert2ByteArray Marshals the states to a []byte
func (sm *States) convert2ByteArray() ([]byte, error) {
	statesData, err := yaml.Marshal(sm)
	log.Debugf("statesData: %s", string(statesData))
	if err != nil {
		return nil, err
	}
	return statesData, err
}

//convert2String Marshals the states to string.
func (sm *States) convert2String() (string, error) {
	statesData, err := sm.convert2ByteArray()
	if err != nil {
		return "", err
	}
	return string(statesData), err
}

//Search a state in states
func (sm *States) _getState(state string) (*State, error) {
	log.Debug("Entering... _getState")
	log.Debugf("Read states=%s\n", state)
	for i := 0; i < len(sm.StateArray); i++ {
		if sm.StateArray[i].Name == state {
			return &sm.StateArray[i], nil
		}
	}
	return nil, errors.New("State: " + state + " not found!")
}

//GetStates returns the list of states with a given status. if the status is an empty string then it returns all states.
func (sm *States) GetStates(status string, extensionsOnly bool, recursive bool, langs []string) (*States, error) {
	log.Debug("Entering... GetStates")
	errStates := sm.readStates()
	if errStates != nil {
		return nil, errStates
	}
	states := &States{
		StateArray:              make([]State, 0),
		StatesPath:              sm.StatesPath,
		ExtensionName:           sm.ExtensionName,
		ExecutedByExtensionName: sm.ExecutedByExtensionName,
		ExecutionID:             sm.ExecutionID,
		StartTime:               sm.StartTime,
		EndTime:                 sm.EndTime,
		Status:                  sm.Status,
		ParentExtensionName:     sm.ParentExtensionName,
		mux:                     &sync.Mutex{},
	}
	statuses, err := sm.CalculateStatesToRun(FirstState, LastState)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(sm.StateArray); i++ {
		// isExention, err := IsExtension(sm.StateArray[i].Name)
		// if err != nil {
		// 	return nil, err
		// }
		if _, ok := statuses[sm.StateArray[i].Name]; ok {
			sm.StateArray[i].NextRun = true
		}
		if sm.StateArray[i].IsExtension {
			//			if strings.HasPrefix(sm.StateArray[i].Script, global.ClientPath+" extension") {
			states.StateArray = append(states.StateArray, sm.StateArray[i])
			if recursive {
				smp, err := GetStatesManager(sm.StateArray[i].Name)
				if err != nil {
					return nil, err
				}
				subStates, err := smp.GetStates(status, extensionsOnly, recursive, langs)
				if err != nil {
					return nil, err
				}
				states.StateArray = append(states.StateArray, subStates.StateArray...)
			}
		} else if !extensionsOnly {
			states.StateArray = append(states.StateArray, sm.StateArray[i])
		}
	}
	//Filter on status
	var resultStates States
	if status != "" {
		for i := 0; i < len(states.StateArray); i++ {
			if states.StateArray[i].Status == status {
				resultStates.StateArray = append(resultStates.StateArray, states.StateArray[i])
			}
		}
	} else {
		resultStates = *states
	}
	//Translate states
	if langs != nil {
		for i := 0; i < len(resultStates.StateArray); i++ {
			resultStates.StateArray[i].Label, _ = i18nUtils.Translate(resultStates.StateArray[i].Label, resultStates.StateArray[i].Label, langs)
		}
	}
	return &resultStates, nil
}

//SetStates Set the current states with a new states. If overwrite is false, then the 2 states will be merged.
//States marked deleted in the new states will be removed for the current states.
func (sm *States) SetStates(states States, overwrite bool) error {
	log.Debug("Entering... SetStates")
	log.Debug("ExtensionPath:" + sm.StatesPath)
	sm.lock()
	defer sm.unlock()
	err := states.topoSort()
	if err != nil {
		return err
	}
	if _, err := os.Stat(sm.StatesPath); os.IsNotExist(err) {
		log.Debug(errors.New("State file " + sm.StatesPath + " doesn't exist"))
		overwrite = true
	} else {
		errStates := sm.readStates()
		if errStates != nil {
			return errStates
		}
	}
	if sm.isRunning() {
		return errors.New("The current state file has a running, action forbidden:" + sm.StatesPath)
	}
	if overwrite {
		newStates, errDelete := sm.removeDeletedStates(states)
		if errDelete != nil {
			return errDelete
		}
		sm.StateArray = newStates.StateArray
		//Do the topo sort
		err = sm.topoSort()
		//Error means cycles
		if err != nil {
			return err
		}
	} else {
		log.Info("Merge new and old States File")
		errMerge := sm.mergeStates(states)
		if errMerge != nil {
			return errMerge
		}
	}
	errStates := sm.writeStates()
	return errStates
}

//SetStatesStatuses Sets the status of states within a range.
func (sm *States) SetStatesStatuses(status string, fromStateName string, fromIncluded bool, toStateName string, toIncluded bool) error {
	log.Debug("Entering... SetStatesStatuses")
	log.Debug("New Status: " + status)
	log.Debug("From state: " + fromStateName)
	log.Debug("From included: " + strconv.FormatBool(fromIncluded))
	log.Debug("To state: " + toStateName)
	log.Debug("To included: " + strconv.FormatBool(toIncluded))
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	var err error
	if status == "" {
		return errors.New("New Status not defined")
	}
	fromIndex := 0
	if fromStateName != "" {
		fromIndex = sm.getStatePosition(fromStateName)
		log.Debug("From Position: " + fromStateName + " index: " + strconv.Itoa(fromIndex))
		if fromIndex == -1 {
			return errors.New(fromStateName + " not found!")
		}
		if !fromIncluded {
			log.Debug("From excluded -1")
			fromIndex++
		}
		log.Debug("From Position excluded: " + fromStateName + " index: " + strconv.Itoa(fromIndex))
	}
	log.Debug("From: " + fromStateName + " index: " + strconv.Itoa(fromIndex))
	toIndex := len(sm.StateArray) - 1
	if toStateName != "" {
		toIndex = sm.getStatePosition(toStateName)
		log.Debug("To Position: " + toStateName + " index: " + strconv.Itoa(toIndex))
		if toIndex == -1 {
			return errors.New(toStateName + " not found!")
		}
		if !toIncluded {
			log.Debug("To excluded -1")
			toIndex--
		}
		log.Debug("To Position excluded: " + toStateName + " index: " + strconv.Itoa(toIndex))
	}
	log.Debug("To: " + toStateName + " index: " + strconv.Itoa(toIndex))
	if fromIndex > toIndex {
		return errors.New("Incorrect ranges or inclusion")
	}
	for index := fromIndex; index <= toIndex; index++ {
		err = sm.SetState(sm.StateArray[index].Name, status, "", "", -1, true)
		if err != nil {
			return err
		}
	}
	return nil
}

//indexState Search index in states array for a given state
func indexState(states []State, stateName string) int {
	for index, stateAux := range states {
		if stateAux.Name == stateName {
			return index
		}
	}
	return -1
}

//getStatePosition Search index in the current statearray
func (sm *States) getStatePosition(stateName string) int {
	return indexState(sm.StateArray, stateName)
}

//Remove deleted states from the current state and new states.
func (sm *States) removeDeletedStates(newStates States) (*States, error) {
	log.Debug("Remove states with delete true")
	//   	copyStates := make([]State, len(sm.StateArray))
	copyNewStates := make([]State, len(newStates.StateArray))
	//Remove the states marked for deletion
	log.Debug("length of sm.StateArray: " + strconv.Itoa(len(sm.StateArray)))
	//log.Debug("length of copyStates: " + strconv.Itoa(len(copyStates)))
	copy(copyNewStates, newStates.StateArray)
	log.Debug("length of newStates.StateArray: " + strconv.Itoa(len(newStates.StateArray)))
	log.Debug("length of copyNewStates: " + strconv.Itoa(len(copyNewStates)))
	//remove the states marked with a deleted "true'"
	//Loop on all new states
	for _, state := range copyNewStates {
		log.Debug("Check state: " + state.Name + " to be deleted " + strconv.FormatBool(state.Deleted))
		//If the new state is marked as to be deleted
		if state.Deleted {
			//Search the same state in the current state file.
			oldIndex := indexState(sm.StateArray, state.Name)
			log.Debug("currentStates located at " + strconv.Itoa(oldIndex))
			//if found then clean current states
			if oldIndex != -1 {
				//if last one then takes the first elements otherwise remove the one in the middle
				if oldIndex == len(sm.StateArray)-1 {
					sm.StateArray = sm.StateArray[:oldIndex]
				} else {
					sm.StateArray = append(sm.StateArray[:oldIndex], sm.StateArray[oldIndex+1:]...)
				}
			}
			newIndex := indexState(newStates.StateArray, state.Name)
			log.Debug("newStates located at " + strconv.Itoa(newIndex))
			//if found then clean newStates
			if newIndex != -1 {
				//Remove the deleted one from the newStates
				if newIndex == len(newStates.StateArray)-1 {
					newStates.StateArray = newStates.StateArray[:newIndex]
				} else {
					newStates.StateArray = append(newStates.StateArray[:newIndex], newStates.StateArray[newIndex+1:]...)
				}
			}
		}
	}
	//Sort the current States
	err := sm.topoSort()
	if err != nil {
		return &newStates, err
	}
	//Sort the newStates
	err = newStates.topoSort()
	if err != nil {
		return &newStates, err
	}
	return &newStates, nil
}

//addNodes create a graph will all nodes.
func (sm *States) addNodes() (*simple.DirectedGraph, map[string]int64, map[int64]*State) {
	newGraph := simple.NewDirectedGraph()
	statesNodesID := make(map[string]int64)
	statesMap := make(map[int64]*State)
	for i := 0; i < len(sm.StateArray); i++ {
		log.Debug("Add Node: " + sm.StateArray[i].Name)
		n := newGraph.NewNode()
		newGraph.AddNode(n)
		statesNodesID[sm.StateArray[i].Name] = n.ID()
		statesMap[n.ID()] = &sm.StateArray[i]
		log.Debug("Old Node added: " + sm.StateArray[i].Name + " with id: " + strconv.FormatInt(n.ID(), 10))
	}
	return newGraph, statesNodesID, statesMap
}

//addEdgesNext Adds the edges listed in the NextState of a given state to the graph
func (sm *States) addEdgesNext(currentState State, newGraph *simple.DirectedGraph, statesNodesID map[string]int64) error {
	ns := newGraph.Node(statesNodesID[currentState.Name])
	log.Debug("CurrentState:" + currentState.Name)
	log.Debugf("NextStates: %+v", currentState.NextStates)
	for _, stateNext := range currentState.NextStates {
		if id, ok := statesNodesID[stateNext]; ok {
			ne := newGraph.Node(id)
			e := newGraph.NewEdge(ns, ne)
			if ne.ID() != ns.ID() {
				newGraph.SetEdge(e)
				log.Debug("Add Egde from existing: " + currentState.Name + " -> " + stateNext)
			} else {
				return errors.New("Add Edge: Current and next state are the same: " + stateNext)
			}
		} else {
			log.Warning("WARNING: The state next " + stateNext + " listed in states_next attribute of " + currentState.Name + " does not exist")
		}
	}
	return nil
}

//addEdgesPrevious Adds the edges listed in the PreviousState of a given state to the graph
func (sm *States) addEdgesPrevious(currentState State, newGraph *simple.DirectedGraph, statesNodesID map[string]int64) error {
	ns := newGraph.Node(statesNodesID[currentState.Name])
	log.Debug("CurrentState:" + currentState.Name)
	log.Debugf("PreviousStates: %+v", currentState.PreviousStates)
	for _, statePrevious := range currentState.PreviousStates {
		if id, ok := statesNodesID[statePrevious]; ok {
			ne := newGraph.Node(id)
			e := newGraph.NewEdge(ne, ns)
			if ne.ID() != ns.ID() {
				newGraph.SetEdge(e)
				log.Debug("Add Egde from existing: " + currentState.Name + " -> " + statePrevious)
			} else {
				return errors.New("Add Edge: Current and next state are the same: " + statePrevious)
			}
		} else {
			log.Warning("WARNING: The state next " + statePrevious + " listed in states_previous attribute of " + currentState.Name + " does not exist")
		}
	}
	return nil
}

//addEdges Adds all edges of the current states to the graph
func (sm *States) addEdges(newGraph *simple.DirectedGraph, statesNodesID map[string]int64) (*simple.DirectedGraph, error) {
	//Add edges based on state sequence in states file.
	for _, state := range sm.StateArray {
		err := sm.addEdgesNext(state, newGraph, statesNodesID)
		if err != nil {
			return nil, err
		}
	}
	return newGraph, nil
}

//generateStatesGraph Generates directed graph with nodes and edgeds defined in the current states
func (sm *States) generateStatesGraph() (*simple.DirectedGraph, map[int64]*State, error) {
	newGraph, statesNodesID, statesMap := sm.addNodes()
	newGraph, err := sm.addEdges(newGraph, statesNodesID)
	return newGraph, statesMap, err
}

//topoSort Do a topological sort of the current states file.
func (sm *States) topoSort() error {
	log.Debug("Entering in... topoSort")
	isNextStateMigrationDone := false
	for index := range sm.StateArray {
		// Remove state in the nextStates which doesn't exist in the state file.
		for indexNext, nextState := range sm.StateArray[index].NextStates {
			if !isNextStateMigrationDone {
				isNextStateMigrationDone = true
			}
			indexState := indexState(sm.StateArray, nextState)
			if indexState == -1 {
				sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates[:indexNext], sm.StateArray[index].NextStates[indexNext+1:]...)
			}
		}
	}
	if !isNextStateMigrationDone {
		sm.setNextStates()
		//		sm.copyStateToRerunToNextStates()
	}
	statesData, _ := sm.convert2String()
	log.Debugf("%s", statesData)
	newGraph, statesMap, err := sm.generateStatesGraph()
	if err != nil {
		return err
	}
	//Do the topo sort
	err = sm.topoSortGraph(newGraph, statesMap)
	statesData, _ = sm.convert2String()
	log.Debug(statesData)
	return err
}

//topoSortGraph Do a topological sort of a directred graph.
//If cycles the error contains the list cycles and nodes.
//It checks also if the StatesToRerun are after the state that defines them.
//It checks also if the PrerequisiteStatesToRerun are before the state that defines them.
func (sm *States) topoSortGraph(graph *simple.DirectedGraph, statesMap map[int64]*State) error {
	sorted, err := topo.Sort(graph)
	//Error means cycles
	if err != nil {
		cycles := searchCycleOnGraph(graph, statesMap)
		return generateCyclesError(cycles)
	}
	//Generate the new states based on the sort
	sm.generateStatesFromGraph(sorted, graph, statesMap)
	err = sm.checkStatesToRerun()
	if err != nil {
		return err
	}
	// err = sm.checkRerunOnRunOfStates()
	// if err != nil {
	// 	return err
	// }
	err = sm.checkPrerequisiteStates()
	return err
}

//checkStatesToRerun Check if the StatesToRerun are after each state that define them.
func (sm *States) checkStatesToRerun() error {
	statesVisited := make(map[string]string, 0)
	for i := 0; i < len(sm.StateArray); i++ {
		if !sm.StateArray[i].Deleted {
			for _, stateName := range sm.StateArray[i].StatesToRerun {
				if _, ok := statesVisited[stateName]; ok {
					return errors.New("State " + sm.StateArray[i].Name + " contains the StatesToRerun element " + stateName + " which is before the state")
				}
			}
			statesVisited[sm.StateArray[i].Name] = sm.StateArray[i].Name
		}
	}
	return nil
}

//checkPrerequisiteStates Check if the PrerequisiteStates are before each state that define them.
func (sm *States) checkPrerequisiteStates() error {
	log.Debug("Entering.... checkPrerequisiteStates")
	statesVisited := make(map[string]string, 0)
	log.Debug("len(sm.StateArray): " + strconv.Itoa(len(sm.StateArray)))
	for i := len(sm.StateArray) - 1; i >= 0; i-- {
		log.Debug("Current State: " + sm.StateArray[i].Name)
		if !sm.StateArray[i].Deleted {
			for _, stateName := range sm.StateArray[i].PrerequisiteStates {
				log.Debug("Current State: " + sm.StateArray[i].Name + " current prereq: " + stateName)
				if _, ok := statesVisited[stateName]; ok {
					return errors.New("State " + sm.StateArray[i].Name + " contains the PrerequisiteStates element " + stateName + " which is after the state")
				}
			}
			statesVisited[sm.StateArray[i].Name] = sm.StateArray[i].Name
		}
	}
	return nil
}

//checkRerunOnRunOfStates Check if the RerunOnRunOfStates are before each state that define them.
func (sm *States) checkRerunOnRunOfStates() error {
	log.Debug("Entering.... checkRerunOnRunOfStates")
	statesVisited := make(map[string]string, 0)
	log.Debug("len(sm.StateArray): " + strconv.Itoa(len(sm.StateArray)))
	for i := len(sm.StateArray) - 1; i >= 0; i-- {
		log.Debug("Current State: " + sm.StateArray[i].Name)
		if !sm.StateArray[i].Deleted {
			for _, stateName := range sm.StateArray[i].RerunOnRunOfStates {
				log.Debug("Current State: " + sm.StateArray[i].Name + " current prereq: " + stateName)
				if _, ok := statesVisited[stateName]; ok {
					return errors.New("State " + sm.StateArray[i].Name + " contains the checkRerunOnRunOfStates element " + stateName + " which is after the state")
				}
			}
			statesVisited[sm.StateArray[i].Name] = sm.StateArray[i].Name
		}
	}
	return nil
}

//searchCycles search cycles in the current states.
func (sm *States) searchCycles() ([]*States, error) {
	newGraph, statesMap, err := sm.generateStatesGraph()
	if err != nil {
		return nil, err
	}
	statesCycles := searchCycleOnGraph(newGraph, statesMap)
	return statesCycles, nil
}

//searchCycleOnGraph Search cycles in a directed graph
func searchCycleOnGraph(graph *simple.DirectedGraph, statesMap map[int64]*State) []*States {
	cycles := topo.DirectedCyclesIn(graph)
	statesCycles := make([]*States, 0, 0)
	if len(cycles) > 0 {
		for i := 0; i < len(cycles); i++ {
			statesCycle := &States{
				StateArray: make([]State, 0),
				StatesPath: "",
			}
			for j := 0; j < len(cycles[i]); j++ {
				statesCycle.StateArray = append(statesCycle.StateArray, *statesMap[cycles[i][j].ID()])
				log.Debugf("%v->", statesMap[cycles[i][j].ID()].Name)
			}
			log.Debugln("")
			statesCycles = append(statesCycles, statesCycle)
		}
	}
	return statesCycles
}

//hasCycles return error if the current states has cycles.
func (sm *States) hasCycles() error {
	cycles, err := sm.searchCycles()
	if err != nil {
		return err
	}
	log.Debugf("sm.StateArray: %v", sm.StateArray)
	return generateCyclesError(cycles)
}

//generateCyclesError Generate an error listing the cycles.
func generateCyclesError(cycles []*States) error {
	errMsg := "Cycles Found:\n"
	if len(cycles) > 0 {
		for i := 0; i < len(cycles); i++ {
			errMsg += "Cycle " + strconv.Itoa(i) + " : "
			for j := 0; j < len(cycles[i].StateArray); j++ {
				state := cycles[i].StateArray[j]
				log.Debugf("%v->", state.Name)
				errMsg += fmt.Sprintf("%v->", state.Name)
			}
			log.Debugln("")
		}
		return errors.New(errMsg)
	}
	return nil
}

//mergeStates Merge 2 states. If a state is in both states (key is the state.Name) then the old state will be overwritten with the new except the
// status, startTime, endTime and Reason.
//States present in the newStates and not in the current states will be added.
//States marked as deleted in the newStates will be deleted in the current states.
//Once the merge done, the final states will be sorted for execution.
func (sm *States) mergeStates(newStates States) error {
	log.Debug("Entering.... mergeStates")
	//Remove the deleted states from the old and new states-file
	cleanedNewStates, err := sm.removeDeletedStates(newStates)
	if err != nil {
		return err
	}
	//If no state are defined use the new provided stateArray, no merge needed.
	if len(sm.StateArray) == 0 {
		sm.StateArray = cleanedNewStates.StateArray
		return nil
	}
	log.Debug("Topology sort")
	//create a graph with all new states
	newGraph, statesNodesID, statesMap := cleanedNewStates.addNodes()
	//Update the existing states to keep their status...
	//Loop on the old states-file
	for i := 0; i < len(sm.StateArray); i++ {
		//if already inserted in the graph then update the state with the current status and other values
		//otherwize insert it as a new node.
		log.Debug("Checking old states: " + sm.StateArray[i].Name)
		if _, ok := statesNodesID[sm.StateArray[i].Name]; ok {
			log.Debug("Update new Node with old status: " + sm.StateArray[i].Name)
			log.Debugf("Before Merge state: %v", sm.StateArray[i])
			state := statesMap[statesNodesID[sm.StateArray[i].Name]]
			log.Debugf("Before Merge state: %v", state)
			state.Status = sm.StateArray[i].Status
			state.StartTime = sm.StateArray[i].StartTime
			state.EndTime = sm.StateArray[i].EndTime
			state.Reason = sm.StateArray[i].Reason
			state.ExecutionID = sm.StateArray[i].ExecutionID
			state.ExecutedByExtensionName = sm.StateArray[i].ExecutedByExtensionName
			log.Debugf("Merged state: %v", sm.StateArray[i])
			log.Debug("New State Node Updated with old status: " + state.Name)
		} else {
			log.Debug("Add old Node: " + sm.StateArray[i].Name)
			n := newGraph.NewNode()
			newGraph.AddNode(n)
			statesNodesID[sm.StateArray[i].Name] = n.ID()
			statesMap[n.ID()] = &sm.StateArray[i]
			log.Debug("Old Node added: " + sm.StateArray[i].Name + " with id: " + strconv.FormatInt(n.ID(), 10))
			err := sm.addEdgesNext(sm.StateArray[i], newGraph, statesNodesID)
			if err != nil {
				return err
			}
			err = sm.addEdgesPrevious(sm.StateArray[i], newGraph, statesNodesID)
			if err != nil {
				return err
			}
		}
	}

	//Add the new states edges
	newGraph, err = cleanedNewStates.addEdges(newGraph, statesNodesID)
	if err != nil {
		return err
	}
	//Add the old states edges
	// newGraph, err = sm.addEdges(newGraph, statesNodesID)
	// if err != nil {
	// 	return err
	// }

	//Print all edges
	for _, edge := range newGraph.Edges() {
		ns := edge.From().ID()
		ne := edge.To().ID()
		log.Debug("Edge: " + strconv.FormatInt(ns, 10) + " -> " + strconv.FormatInt(ne, 10))
		edgeString := fmt.Sprintf("%s -> %s\n", statesMap[ns].Name, statesMap[ne].Name)
		log.Debug("Edge: " + edgeString)
	}

	//Do the topo sort
	err = sm.topoSortGraph(newGraph, statesMap)
	//Error means cycles
	if err != nil {
		return err
	}
	return nil
}

//generateStatesFromGraph Generate a states from a directed graph.
func (sm *States) generateStatesFromGraph(sorted []graph.Node, graph *simple.DirectedGraph, statesMap map[int64]*State) {
	//Generate the new states based on the sort
	sm.StateArray = make([]State, 0)
	for i := 0; i < len(sorted); i++ {
		log.Debugf("%s|", strconv.FormatInt(sorted[i].ID(), 10))
		log.Debugf("%s|", statesMap[sorted[i].ID()].Name)
		statesMap[sorted[i].ID()].NextStates = make([]string, 0)
		for _, node := range graph.From(sorted[i].ID()) {
			statesMap[sorted[i].ID()].NextStates = append(statesMap[sorted[i].ID()].NextStates, statesMap[node.ID()].Name)
		}
		statesMap[sorted[i].ID()].PreviousStates = make([]string, 0)
		for _, node := range graph.To(sorted[i].ID()) {
			statesMap[sorted[i].ID()].PreviousStates = append(statesMap[sorted[i].ID()].PreviousStates, statesMap[node.ID()].Name)
		}
		sm.StateArray = append(sm.StateArray, *statesMap[sorted[i].ID()])
	}
}

//isInNextState Check if a state is in the NextState of a given state.
func (sm *States) isInNextStates(currentState State, stateName string) bool {
	for _, nextStateName := range currentState.NextStates {
		if nextStateName == stateName {
			return true
		}
	}
	return false
}

//NextStatesIndexOf return the index of a stateName in the NextStates
func (sm *States) NextStatesIndexOf(currentState State, stateName string) int {
	for index, nextStateName := range currentState.NextStates {
		if nextStateName == stateName {
			return index
		}
	}
	return -1
}

//isInPreviousState Check if a state is in the PreviousState of a given state.
func (sm *States) isInPreviousStates(currentState State, stateName string) bool {
	for _, previousStateNAme := range currentState.PreviousStates {
		if previousStateNAme == stateName {
			return true
		}
	}
	return false
}

//IsRunning Check if states is running in the persisted states
func (sm *States) IsRunning() (bool, error) {
	errStates := sm.readStates()
	if errStates != nil {
		return false, errStates
	}
	//Check if states running
	if sm.isRunning() {
		return true, nil
	}
	return false, nil
}

//Check if a status is Running in the current states
func (sm *States) isResetRunning() bool {
	for i := 0; i < len(sm.StateArray); i++ {
		state := sm.StateArray[i]
		if state.Status == StateRUNNING {
			return true
		}
	}
	return false
}

//Check if states engine is running
func (sm *States) isRunning() bool {
	return sm.Status == StateRUNNING
}

//setStateStatus Set the status of a given states. I
//f recusively is true and if the state is an extension then the states of the extension will be set to the status and this recursively.
func (sm *States) setStateStatus(state State, status string, recursively bool) error {
	log.Debug("Entering.... setStateStatus state:" + state.Name + " status:" + status + " recursively:" + strconv.FormatBool(recursively))
	index := indexState(sm.StateArray, state.Name)
	if index == -1 {
		return errors.New("State: " + state.Name + " not found!")
	}
	if state.Status != StateSKIP {
		log.Debugln("Change status of " + state.Name + " to " + status)
		sm.StateArray[index].Status = status
	}
	sm.StateArray[index].StartTime = ""
	sm.StateArray[index].EndTime = ""
	sm.StateArray[index].Reason = ""
	if recursively && state.IsExtension {
		log.Debug(state.Name + " is an extension")
		extensionStateManager, err := GetStatesManager(state.Name)
		if err != nil {
			return err
		}
		err = extensionStateManager.readStates()
		if err != nil {
			return err
		}
		for _, state := range extensionStateManager.StateArray {
			err := extensionStateManager.setStateStatus(state, status, true)
			if err != nil {
				return err
			}
		}
	}
	errStates := sm.writeStates()
	if errStates != nil {
		return errStates
	}
	//if the state is set to ready and is part of an extension inserted in another extension
	//then the caller state in the parent extension must be set to ready too
	if sm.ParentExtensionName != "" && status == StateREADY {
		err := sm.setParentStateStatus(StateREADY)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sm *States) setParentStateStatus(status string) error {
	//No parent
	if sm.ParentExtensionName == "" {
		return nil
	}
	//Update the parent extension name in the inserted extension state-file
	stateManager, errStateManager := GetStatesManager(sm.ParentExtensionName)
	if errStateManager != nil {
		logger.AddCallerField().Error(errStateManager.Error())
		return errStateManager
	}
	err := stateManager.readStates()
	if err != nil {
		return err
	}
	stateFound, errState := stateManager._getState(sm.ExtensionName)
	if errState != nil {
		return errState
	}
	if stateFound.Status != status {
		stateFound.Status = status
		stateFound.StartTime = sm.StartTime
		stateFound.EndTime = sm.EndTime
		err = stateManager.writeStates()
		if err != nil {
			return err
		}
		if stateManager.ParentExtensionName != "" {
			return stateManager.setParentStateStatus(status)
		}
	}
	return nil
}

func SetMock(mock bool) {
	global.Mock = mock
}

//Retrieve level
func GetMock() bool {
	return global.Mock
}

//ResetEngine Reset states, all non-skip state will be set to READY recursively
//No RUNNING state must be found.
func (sm *States) ResetEngine() error {
	log.Debug("Entering... ResetEngine")
	sm.lock()
	defer sm.unlock()
	//Read states
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	log.Debug("ResetEngine... states has been read")
	//Check if states running
	if sm.isResetRunning() {
		err := errors.New("Deployment is running, can not proceed")
		log.Debug(err.Error())
		return err
	}
	//Reset states
	for _, state := range sm.StateArray {
		err := sm.setStateStatus(state, StateREADY, true)
		if err != nil {
			return err
		}
	}
	sm.StartTime = ""
	sm.EndTime = ""
	sm.Status = ""
	//Write states
	errStates = sm.writeStates()
	return errStates
}

//ResetEngine Reset execution Info. This is not recursive yet.
//No RUNNING state must be found.
func (sm *States) ResetEngineExecutionInfo() error {
	log.Debug("Entering... ResetEngineExecutionInfo")
	sm.lock()
	defer sm.unlock()
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	log.Debug("ResetEngine... states has been read")
	//Check if states running
	if sm.isRunning() {
		err := errors.New("Deployment is running, can not proceed")
		log.Debug(err.Error())
		return err
	}
	sm.ExecutedByExtensionName = ""
	sm.ExecutionID = 0
	for i := 0; i < len(sm.StateArray); i++ {
		sm.StateArray[i].ExecutedByExtensionName = ""
		sm.StateArray[i].ExecutionID = 0
	}
	errStates = sm.writeStates()
	return errStates
}

//GetState return a state providing its name
func (sm *States) GetState(state string, langs []string) (*State, error) {
	log.Debug("Entering... GetState")
	log.Debugf("Read states=%s\n", state)
	errStates := sm.readStates()
	if errStates != nil {
		return nil, errStates
	}
	stateFound, errState := sm._getState(state)
	if stateFound != nil && errState != nil && langs != nil {
		stateFound.Label, _ = i18nUtils.Translate(stateFound.Label, stateFound.Label, langs)
	}
	return stateFound, errState
}

//SetState Set a state status
func (sm *States) SetState(state string, status string, reason string, script string, scriptTimout int, recursivelly bool) error {
	log.Debugf("Read states=%s\n", state)
	sm.lock()
	defer sm.unlock()
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	stateFound, errState := sm._getState(state)
	if errState != nil {
		return errState
	}
	log.Debugf("Set state %s with status %s", stateFound.Name, status)
	if status != "" {
		if status != StateFAILED &&
			status != StateREADY &&
			status != StateRUNNING &&
			status != StateSKIP &&
			status != StateSUCCEEDED {
			return errors.New("Invalid status:" + status + " (" + StateREADY + "," + StateSKIP + "," + StateRUNNING + "," + StateSUCCEEDED + "," + StateFAILED + ")")
		}
	}
	//if status READY go recursivelly
	if status == StateREADY {
		err := sm.setStateStatus(*stateFound, status, recursivelly)
		if err != nil {
			return err
		}
	}
	stateFound.Status = status
	if status == StateFAILED {
		stateFound.Reason = reason
	} else {
		stateFound.Reason = ""
	}
	if script != "" {
		stateFound.Script = script
	}
	if scriptTimout != -1 {
		stateFound.ScriptTimeout = scriptTimout
	}
	errWriteStates := sm.writeStates()
	if errWriteStates != nil {
		return errWriteStates
	}
	return nil
}

func (sm *States) setExecutionID(state string, callerState *State) (*State, error) {
	log.Debugf("Entering in... setExecutionID")
	stateFound, errState := sm._getState(state)
	if errState != nil {
		return nil, errState
	}
	if callerState == nil {
		log.Debug("CallerState is nil for running state " + state)
		stateFound.ExecutedByExtensionName = sm.ExecutedByExtensionName
		stateFound.ExecutionID = sm.ExecutionID
	} else {
		log.Debug("CallerState is " + callerState.Name + " for running state " + state)
		stateFound.ExecutedByExtensionName = callerState.ExecutedByExtensionName
		stateFound.ExecutionID = callerState.ExecutionID
	}
	errWriteStates := sm.writeStates()
	if errWriteStates != nil {
		return nil, errWriteStates
	}
	return stateFound, nil
}

//setStateStatusWithTimeStamp Set a state status with timestamp
func (sm *States) setStateStatusWithTimeStamp(isStart bool, state string, status string, reason string) error {
	log.Debugf("Entering in... setStateStatusWithTimeStamp")
	stateFound, errState := sm._getState(state)
	if errState != nil {
		return errState
	}
	log.Debugf("Set state %s with status %s", stateFound.Name, status)
	stateFound.Status = status
	if status == StateFAILED {
		stateFound.Reason = reason
		// err := sm.setDependencyStatus(isStart, state, status, "being a dependency of the failed state "+state)
		// if err != nil {
		// 	return err
		// }
	} else {
		stateFound.Reason = ""
	}
	log.Debug("ExecutedByExtensionName: " + stateFound.ExecutedByExtensionName)
	log.Debug("ExecutionID: " + strconv.Itoa(stateFound.ExecutionID))
	timeNow := time.Now().UTC().Format(time.UnixDate)
	if isStart {
		stateFound.StartTime = timeNow
		stateFound.EndTime = ""
	} else {
		stateFound.EndTime = timeNow
	}
	errWriteStates := sm.writeStates()
	if errWriteStates != nil {
		return errWriteStates
	}
	return nil
}

func (sm *States) InsertStateFromExtensionName(extensionName string, pos int, stateName string, before bool, overwrite bool) error {
	log.Debug("Entering..... InsertStateFromExtensionName")
	log.Debug("State name: " + stateName)
	log.Debug("State position: " + strconv.Itoa(pos))
	log.Debug("State before: " + strconv.FormatBool(before))
	manifestPath, err := GetRegisteredExtensionPath(extensionName)
	if err != nil {
		return err
	}
	manifestPath = filepath.Join(manifestPath, global.DefaultExtenstionManifestFile)
	log.Debug("manifestPath: " + manifestPath)
	manifestBytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	cfg, err := config.ParseYaml(string(manifestBytes))
	if err != nil {
		return err
	}
	stateCfg, err := cfg.Get("call_state")
	if err != nil {
		err = cfg.Set("call_state.name", extensionName)
		if err != nil {
			return err
		}
		stateCfg, _ = cfg.Get("call_state")
	} else {
		err = stateCfg.Set("name", extensionName)
	}
	if err != nil {
		return err
	}
	err = stateCfg.Set("is_extension", true)
	if err != nil {
		return err
	}
	stateString, err := config.RenderYaml(stateCfg.Root)
	if err != nil {
		return err
	}
	log.Debug("call_state: " + stateString)
	err = sm.InsertStateFromString(stateString, pos, stateName, before, overwrite)
	if err != nil {
		return err
	}
	return nil
}

//InsertStateFromString Insert state at a given position, before or after a given state.
//The state Def is provided as a string
//If the position is 0 and the stateName is not provided then the state will be inserted taking into account the PreviousStates and NextStates of the inserted state.
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) InsertStateFromString(stateDef string, pos int, stateName string, before bool, overwrite bool) error {
	log.Debug("Entering..... InsertStateFromString")
	log.Debug("State name: " + stateName)
	log.Debug("State position: " + strconv.Itoa(pos))
	log.Debug("State before: " + strconv.FormatBool(before))

	var stateAux State
	err := yaml.Unmarshal([]byte(stateDef), &stateAux)
	if err != nil {
		return err
	}
	return sm.InsertState(stateAux, pos, stateName, before, overwrite)
}

//InsertState Insert state at a given state position, before or after a given state.
//If the referencePosition is not 0, then that position will be used as reference for insertion
//If the referencePosition is equal 0 and the referenceStateName is provided, the position of the referenceStateName will be used as reference for insertion
//If the referencePosition is 0 and the referenceStateName is not provided then the state will be inserted taking into account the PreviousStates and NextStates of the inserted state.
//If the state is already present, then it is updated if overwrite is true
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) InsertState(state State, referencePosition int, referenceStateName string, before bool, overwrite bool) error {
	log.Debug("Entering..... InsertState")
	log.Debug("Reference State name: " + referenceStateName)
	log.Debugf("State to be inserted: %v", state)
	sm.lock()
	defer sm.unlock()
	errStates := sm.readStates()
	if errStates != nil {
		logger.AddCallerField().Error(errStates.Error())
		return errStates
	}
	if sm.isRunning() {
		return errors.New("Insert can not be executed while a deployment is running")
	}
	if state.Name == "" || (!state.IsExtension && state.Script == "") {
		return errors.New("The state name or script is missing")
	}
	mustUpdateCallerState := false
	existingPosition := sm.getStatePosition(state.Name)
	if existingPosition != -1 {
		if overwrite {
			mustUpdateCallerState = true
		} else {
			return errors.New("State name " + state.Name + " already exists")
		}
	}
	// valid, err := IsExtension(state.Name)
	// if err != nil {
	// 	logger.AddCallerField().Error(err.Error())
	// 	return err
	// }
	//We are inserting an extension and so the extension must be registered
	if state.IsExtension {
		registered := IsExtensionRegistered(state.Name)
		if !registered {
			return errors.New("The state name " + state.Name + " is not registered")
		}
	}
	//Set the position with the provided position
	whereToInsertedPosition := referencePosition
	//if the stateName is provided then the insertion will be relative to that state
	//Search the position of that state as new position
	if referenceStateName != "" {
		whereToInsertedPosition = sm.getStatePosition(referenceStateName)
		if whereToInsertedPosition == -1 {
			err := errors.New("state : " + referenceStateName + " not found")
			logger.AddCallerField().Error(err.Error())
			return err
		}
		whereToInsertedPosition++
		log.Debug("Position:" + strconv.Itoa(whereToInsertedPosition))
	} else if whereToInsertedPosition == 0 && (len(state.NextStates) == 0 || len(state.PreviousStates) == 0) {
		return errors.New("The position, state name and previous and next states are undefined")
	} else if whereToInsertedPosition != 0 && (whereToInsertedPosition < 1 || whereToInsertedPosition > len(sm.StateArray)) {
		return errors.New("The position must be between 1 and " + strconv.Itoa(len(sm.StateArray)) + " currently:" + strconv.Itoa(whereToInsertedPosition))
	}

	//Copy the state at the end but it will be overwritten by the copy
	bckStateArray := make([]State, 0)
	bckStateArray = append(bckStateArray, sm.StateArray...)
	log.Debugf("%v", bckStateArray)
	//The new state doesn't provide relative positionning
	if len(state.NextStates) == 0 && len(state.PreviousStates) == 0 {
		//if state already inserted, then just update the state
		if mustUpdateCallerState {
			updateState(&sm.StateArray[existingPosition], state)
		} else {
			arrayPos := whereToInsertedPosition
			//Do the insertion based on position
			if before {
				arrayPos = whereToInsertedPosition - 1
			}
			//Update the PreviousStates and NextStates surrounding states
			if arrayPos > 0 {
				if !sm.isInNextStates(sm.StateArray[arrayPos-1], state.Name) {
					sm.StateArray[arrayPos-1].NextStates = append(sm.StateArray[arrayPos-1].NextStates, state.Name)
				}
			}
			if arrayPos < len(sm.StateArray) {
				if !sm.isInNextStates(state, sm.StateArray[arrayPos].Name) {
					state.NextStates = append(state.NextStates, sm.StateArray[arrayPos].Name)
				}
			}
		}
	} else {
		//if state already inserted update the current state and delete the NextStates reference
		// referenced in PreviousStats of the current state
		if mustUpdateCallerState {
			log.Debug("Not position insertion")
			for _, stateName := range sm.StateArray[existingPosition].PreviousStates {
				log.Debugf("Cleaning Referenced NextStates of %s", stateName)
				statePos := sm.getStatePosition(stateName)
				if statePos == -1 {
					err := errors.New("state : " + stateName + " not found")
					logger.AddCallerField().Error(err.Error())
					return err
				}
				//Search the index of the state in the NextStates
				index := sm.NextStatesIndexOf(sm.StateArray[statePos], sm.StateArray[existingPosition].Name)
				if index != -1 {
					//remove entry in NextStates
					log.Debugf("Remove NextStates %s from %s", sm.StateArray[existingPosition].Name, sm.StateArray[statePos].Name)
					sm.StateArray[statePos].NextStates = append(sm.StateArray[statePos].NextStates[:index], sm.StateArray[statePos].NextStates[index+1:]...)
				}
			}
			updateState(&sm.StateArray[existingPosition], state)
			log.Debugf("State after update: %v", sm.StateArray[existingPosition])
		}
		//As the topology sort is a based on the NextStates, we need to
		//update the NextStates of the states referenced by the PreviousStates of provided state.
		for _, stateName := range state.PreviousStates {
			statePos := sm.getStatePosition(stateName)
			if statePos == -1 {
				err := errors.New("state : " + stateName + " not found")
				logger.AddCallerField().Error(err.Error())
				return err
			}
			if !sm.isInNextStates(sm.StateArray[statePos], stateName) {
				sm.StateArray[statePos].NextStates = append(sm.StateArray[statePos].NextStates, state.Name)
			}
		}
	}
	if !mustUpdateCallerState {
		sm.StateArray = append(sm.StateArray, state)
		err := sm.setStateStatus(state, StateREADY, true)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return err
		}
	}
	err := sm.topoSort() //	err = sm.hasCycles()
	if err != nil {
		log.Debugf("bckStateArray: %v", bckStateArray)
		sm.StateArray = make([]State, 0)
		sm.StateArray = append(sm.StateArray, bckStateArray...)
		logger.AddCallerField().Error(err.Error())
		return errors.New(err.Error())
	}
	sm.setDefaultValues()
	err = sm.writeStates()
	if err != nil {
		return err
	}
	//We are inserting an extension
	if state.IsExtension {
		//Update the parent extension name in the inserted extension state-file
		stateManager, errStateManager := GetStatesManager(state.Name)
		if errStateManager != nil {
			logger.AddCallerField().Error(errStateManager.Error())
			return errStateManager
		}
		err = stateManager.readStates()
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return err
		}
		stateManager.ParentExtensionName = sm.ExtensionName
		log.Debug("Parent Extension name:" + sm.ExtensionName)
		return stateManager.writeStates()
	}
	return nil

}

func updateState(currentState *State, newState State) {
	currentState.Phase = newState.Phase
	currentState.Label = newState.Label
	currentState.LogPath = newState.LogPath
	currentState.Script = newState.Script
	currentState.ScriptTimeout = newState.ScriptTimeout
	currentState.Protected = newState.Protected
	currentState.Deleted = newState.Deleted
	currentState.PrerequisiteStates = newState.PrerequisiteStates
	currentState.StatesToRerun = newState.StatesToRerun
	currentState.RerunOnRunOfStates = newState.RerunOnRunOfStates
	currentState.PreviousStates = newState.PreviousStates
	currentState.NextStates = newState.NextStates
	currentState.IsExtension = newState.IsExtension
}

//updateCallers update the caller state already inserted in an another extension states-file
func updateCallers(extensionName string) error {
	log.Debug("Entering... updateCallers: " + extensionName)
	//Search manager
	stateManager, errStateManager := GetStatesManager(extensionName)
	if errStateManager != nil {
		logger.AddCallerField().Error(errStateManager.Error())
		return errStateManager
	}
	err := stateManager.readStates()
	if err != nil {
		return err
	}

	//If no parent nothing to do
	log.Debug("stateManager.ParentExtensionName:" + stateManager.ParentExtensionName)
	if stateManager.ParentExtensionName != "" {
		stateParentManager, errStateManager := GetStatesManager(stateManager.ParentExtensionName)
		if errStateManager != nil {
			logger.AddCallerField().Error(errStateManager.Error())
			return errStateManager
		}
		err := stateParentManager.readStates()
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return err
		}
		//Search the caller state in the parent states-file
		currentPosition := stateParentManager.getStatePosition(extensionName)
		log.Debug("updateCallers.curcurrentPositionrent: " + strconv.Itoa(currentPosition))
		if currentPosition == -1 {
			currentPosition = 0
		}
		err = stateParentManager.InsertStateFromExtensionName(extensionName, currentPosition, "", true, true)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return err
		}
	}
	return nil
}

//DeleteState Delete a state at a given position or with a given name
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) DeleteState(pos int, stateName string) error {
	log.Debug("Entering..... DeleteState")
	sm.lock()
	defer sm.unlock()
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	if sm.isRunning() {
		return errors.New("Insert can not be executed while a deployment is running")
	}
	position := pos
	var err error
	if stateName != "" {
		position = sm.getStatePosition(stateName)
		if position == -1 {
			return errors.New("state :" + stateName + " not found")
		}
		position++
	}
	log.Debug("Position:" + strconv.Itoa(position))
	if position < 1 || position > len(sm.StateArray) {
		return errors.New("The position must be between 1 and " + strconv.Itoa(len(sm.StateArray)) + " currently:" + strconv.Itoa(position))
	}
	arrayPos := position - 1
	log.Debug("Protected:" + strconv.FormatBool(sm.StateArray[arrayPos].Protected))
	if sm.StateArray[arrayPos].Protected {
		return errors.New("The state " + sm.StateArray[arrayPos].Name + " is protected and can not be deleted")
	}
	stateNameAux := sm.StateArray[arrayPos].Name
	isExtension := sm.StateArray[arrayPos].IsExtension
	copy(sm.StateArray[arrayPos:], sm.StateArray[arrayPos+1:])
	sm.StateArray[len(sm.StateArray)-1] = State{} // or the zero value of T
	sm.StateArray = sm.StateArray[:len(sm.StateArray)-1]
	err = sm.topoSort()
	if err != nil {
		return err
	}
	err = sm.writeStates()
	if err != nil {
		return err
	}
	//Update the parent extension name in the inserted extension state-file
	// valid, err := IsExtension(stateNameAux)
	// if err != nil {
	// 	log.Debug(err.Error())
	// 	return err
	// }
	if isExtension {
		stateManager, errStateManager := GetStatesManager(stateNameAux)
		if errStateManager != nil {
			logger.AddCallerField().Error(errStateManager.Error())
			return errStateManager
		}
		err = stateManager.readStates()
		if err != nil {
			return err
		}
		stateManager.ParentExtensionName = ""
		return stateManager.writeStates()
	}
	return nil
}

//getLogPath Search logPath for a given state in states structure
func (sm *States) getLogPath(state string) (string, error) {
	stateFound, err := sm._getState(state)
	if err != nil {
		return "", err
	}
	logPath := stateFound.LogPath
	if logPath == "" {
		return logPath, errors.New("No logPath available for " + state)
	}
	log.Debugf("LogPath:%s", logPath)
	return logPath, nil
}

//GetLogs Get logs from a given position, a given length. The length is the number of characters to return if bychar is true otherwize is the number of lines.
func (sm *States) GetLogs(position int64, length int64, bychar bool) (string, error) {
	var data []byte
	states, err := sm.GetStates("", false, false, nil)
	if err != nil {
		return string(data), err
	}

	for _, state := range states.StateArray {
		bytes, err := sm.GetLog(state.Name, 0, math.MaxInt64, bychar)
		if err != nil {
			return string(data), err
		}
		data = append(data, bytes...)
	}
	return string(data), nil
}

/*GetLog Retrieve log of a given state.
state: Look at the log of a given state.
position: start at position (byte) in the log (default:0)
len: number of byte to retrieve.
Return a []byte containing the requested log or an error.
*/
func (sm *States) GetLog(state string, position int64, length int64, bychar bool) ([]byte, error) {
	log.Debugf("statesPath=%s", sm.StatesPath)
	log.Debugf("state=%s\n", state)
	//log.Debugf("position:%s", strconv.FormatInt(position, 10))
	//log.Debugf("length:%s", strconv.FormatInt(length, 10))
	var logFile *os.File
	var logPath string
	var data []byte
	if state == "" && sm.ExtensionName != "mock" && sm.ExtensionName != "cr" {
		return nil, errors.New("State must be defined")
	}
	//Mock log
	switch sm.ExtensionName {
	case "mock":
		logFile, _ := ioutil.TempFile("", "mock")
		defer os.Remove(logFile.Name()) // clean up
		var buffer bytes.Buffer
		for i := 1; i <= 200; i++ {
			buffer.WriteString("Line " + strconv.Itoa(i) + " Mock log line\n")
		}
		logString := buffer.String()
		logFile.WriteString(logString)
		logFile.Close()
		logPath = logFile.Name()
	case "cr":
		var err error
		if position == 0 {
			//copy to temp file because log.Debug... continue to fill the actual commands-runner.log
			crLogTempFile, err = ioutil.TempFile("/tmp/", "/commands-runner-log")
			if err != nil {
				return nil, err
			}
			logFile, err := os.Open(logger.LogFile.Filename)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(crLogTempFile, logFile)
			if err != nil {
				return nil, err
			}
		}
		logPath = crLogTempFile.Name()
	default:
		// read log
		errStates := sm.readStates()
		if errStates != nil {
			return nil, errStates
		}
		//Search for the path of the log in the states
		logPathFound, errPath := sm.getLogPath(state)
		if errPath != nil {
			return nil, errPath
		}
		dir := filepath.Dir(logPathFound)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(sm.StatesPath), dir)
		}
		//Check if log exists and rename it for backup
		logPath = filepath.Join(dir, filepath.Base(logPathFound))
	}
	//Open the log file.
	var errFile error
	log.Debug("LogfilePath:" + logPath)
	logFile, errFile = os.Open(logPath)
	if errFile != nil {
		return nil, errFile
	}
	//Search last position
	lastPosition := position + length
	//check of overflow in case of the length is not defined and thus set to math.MaxInt64
	if lastPosition < 0 {
		lastPosition = math.MaxInt64
	}
	if !bychar {
		// Retrieve the log data by lines
		scanner := bufio.NewScanner(logFile)
		var pos = int64(1)

		var logDataArray []string
		//scan the file by lines
		for scanner.Scan() {
			log.Debugf("Line NB:%d", pos)
			if pos >= lastPosition {
				break
			}
			if pos >= position {
				logDataArray = append(logDataArray, scanner.Text())
			}
			pos++
		}
		if errLogData := scanner.Err(); errLogData != nil {
			return nil, errLogData
		}
		logData := strings.Join(logDataArray[:], "\n")
		data = []byte(logData)
	} else {
		//Search last byte position in file to retrieve
		fi, errStat := logFile.Stat()
		if errStat != nil {
			return nil, errStat
		}
		fileSize := fi.Size()
		if lastPosition > fileSize {
			lastPosition = fileSize
		}
		//log.Debugf("lastPosition:%s", strconv.FormatInt(lastPosition, 10))
		//log.Debugf("position:%s", strconv.FormatInt(position, 10))
		if position < lastPosition {
			data = make([]byte, lastPosition-position)
			nb, _ := logFile.ReadAt(data, position)
			log.Debug("Nb Bytes read:" + strconv.Itoa(nb))
			data = data[:nb]
		} else {
			data = []byte("")
		}
	}
	return data, nil
}

//Start states from beginning to end
func (sm *States) Start() error {
	log.Debug("Enterring... Start")
	return sm.Execute(FirstState, LastState, nil, nil)
}

func (sm *States) CalculateStatesToRun(fromState string, toState string) (map[string]string, error) {
	log.Debug("Enterring... calculateStatesToRun from " + fromState + " to " + toState)
	log.Debug("State:" + sm.StatesPath)
	log.Debug("From state:" + fromState)
	log.Debug("To   state:" + toState)
	statuses := make(map[string]string, 0)
	err := sm.setCalculatedStatesToRerun()
	if err != nil {
		return statuses, err
	}
	statesVisited := make(map[string]string, 0)
	statesToProcess := make([]State, 0)
	//Search all READY or FAILED states and populate statesToPRocess
	toExecute := false || fromState == FirstState
	for _, state := range sm.StateArray {
		//Start using state when reach the fromState
		toExecute = toExecute || state.Name == fromState
		if toExecute &&
			(state.Status == StateREADY ||
				state.Status == StateFAILED ||
				(state.Status != StateSKIP && state.Phase == PhaseAtEachRun)) {
			statesToProcess = append(statesToProcess, state)
			statuses[state.Name] = state.Status
		}
		//Stop when we processed to the toState
		if state.Name == toState {
			break
		}
	}
	log.Debugf("statesToProcess: %+v", statesToProcess)
	//Until no states to process
	for len(statesToProcess) != 0 {
		currentState := statesToProcess[0]
		log.Debugf("Current state: %v", currentState)
		//Skip if state already visited
		if _, ok := statesVisited[currentState.Name]; !ok {
			for _, stateName := range currentState.CalculatedStatesToRerun {
				state, err := sm._getState(stateName)
				if err != nil {
					return nil, errors.New("The state " + stateName + " is not an existing state")
				}
				if state.Status != StateSKIP {
					statuses[stateName] = StateREADY
					statesToProcess = append(statesToProcess, *state)
				}
			}
			//Mark state as visited
			statesVisited[currentState.Name] = currentState.Name
		}
		//AS state get processed then remove it from the list
		statesToProcess = statesToProcess[1:]
	}
	return statuses, nil
}

func (sm *States) setCalculatedStatesToRun(statuses map[string]string) error {
	log.Debug("Enterring... setCalculatedStatus")
	for stateName, status := range statuses {
		state, err := sm._getState(stateName)
		if err != nil {
			return errors.New("The state " + stateName + " is not an existing state. error=" + err.Error())
		}
		err = sm.setStateStatus(*state, status, true)
		if err != nil {
			return errors.New("Can not set " + stateName + " to status " + status + " recursively. error=" + err.Error())
		}
		//		state.Status = status
	}
	return nil
}

func (sm *States) preprocessingExecute(fromState string, toState string) error {
	errStates := sm.readStates()
	if errStates != nil {
		log.Debug(errStates.Error())
		return errStates
	}
	if sm.isRunning() {
		err := errors.New("Already running")
		log.Debug(err.Error())
		return err
	}
	//check for cycles
	err := sm.topoSort()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	//Calculate statuses
	statuses, err := sm.CalculateStatesToRun(fromState, toState)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	//Update statuses
	err = sm.setCalculatedStatesToRun(statuses)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	return nil
}

func (sm *States) setStatesExecutionID(callerState *State) error {
	errStates := sm.readStates()
	if errStates != nil {
		log.Debug(errStates.Error())
		return errStates
	}
	if sm.isRunning() {
		err := errors.New("Already running")
		log.Debug(err.Error())
		return err
	}
	if callerState == nil {
		sm.ExecutedByExtensionName = sm.ExtensionName
		sm.ExecutionID = sm.ExecutionID + 1
	} else {
		sm.ExecutedByExtensionName = callerState.ExecutedByExtensionName
		sm.ExecutionID = callerState.ExecutionID
	}
	sm.Status = StatePREPROCESSING
	errStates = sm.writeStates()
	if errStates != nil {
		log.Debug(errStates.Error())
		return errStates
	}
	return nil
}

func (sm *States) setExecutionTimesAndStatesStatus(status string, callerState *State) error {
	errStates := sm.readStates()
	if errStates != nil {
		log.Debug(errStates.Error())
		return errStates
	}
	timeNow := time.Now().UTC().Format(time.UnixDate)
	if status == StateRUNNING {
		sm.StartTime = timeNow
		sm.EndTime = ""
	} else {
		sm.EndTime = timeNow
	}
	sm.Status = status
	// if this execution was called from a parent states file
	// then no need to set the parent status as the parent status is set during
	// the parent execution... also if that test was not done, the parent states status
	// could flip from READY, RUNNING, SUCCEEDED each time a sub-process is called from the parent.
	if callerState == nil {
		//Set the caller state status and timestamp
		errStates = sm.setParentStateStatus(status)
		if errStates != nil {
			log.Debug(errStates.Error())
			return errStates
		}
	}
	errStates = sm.writeStates()
	if errStates != nil {
		log.Debug(errStates.Error())
		return errStates
	}
	return nil
}

//Execute states from state 'fromState' to state 'toState'
func (sm *States) Execute(fromState string, toState string, callerState *State, callerOutFile *os.File) error {
	if callerState == nil {
		err := logger.LogFile.Rotate()
		if err != nil {
			log.Error(err.Error())
		}
		log.Infof("Execute %s from %s to %s", sm.ExtensionName, fromState, toState)
	}
	log.Debug("Enterring... Execute from " + fromState + " to " + toState)
	log.Debug("State:" + sm.StatesPath)
	log.Debug("From state:" + fromState)
	log.Debug("To   state:" + toState)
	err := sm.setStatesExecutionID(callerState)
	//Init executionBy and ID, for the time being it is set to the sm.ExecutionName but later it could be set to the calling extension name
	err = sm.preprocessingExecute(fromState, toState)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	errStartTime := sm.setExecutionTimesAndStatesStatus(StateRUNNING, callerState)
	if errStartTime != nil {
		log.Debug(errStartTime.Error())
		return errStartTime
	}
	err = sm.executeStates(fromState, toState, callerState, callerOutFile)
	status := StateSUCCEEDED
	if err != nil {
		status = StateFAILED
	}
	errStopTime := sm.setExecutionTimesAndStatesStatus(status, callerState)
	if errStopTime != nil {
		log.Debug(errStopTime.Error())
		return errStopTime
	}
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	return nil
}

//Execute states
func (sm *States) executeStates(fromState string, toState string, callerState *State, callerOutFile *os.File) error {
	// if callerState != nil {
	// 	sm.ExecutedByExtensionName = callerState.ExecutedByExtensionName
	// 	sm.ExecutionID = callerState.ExecutionID
	// }
	toExecute := false || fromState == FirstState
	for i := 0; i < len(sm.StateArray); i++ {
		state := sm.StateArray[i]
		//Executed in panic case
		defer func() {
			if r := recover(); r != nil {
				log.Debug(r)
				sm.setStateStatusWithTimeStamp(false, state.Name, StateFAILED, "Panic Error, check logs")
			}
		}()
		log.Debug("Processing state:" + state.Name)
		if state.Name == fromState {
			toExecute = true
		}
		log.Debug("To execute:" + strconv.FormatBool(toExecute))
		//Set to Ready to rerun PhaseAtEachRun state.
		if state.Phase == PhaseAtEachRun && state.Status != StateSKIP {
			errSetReady := sm.setStateStatusWithTimeStamp(true, state.Name, StateREADY, "")
			if errSetReady != nil {
				log.Debug(errSetReady.Error())
				return errSetReady
			}
			state.Status = StateREADY
		}
		if state.Status == StateSUCCEEDED || state.Status == StateSKIP {
			log.Debug("Skip:" + state.Name)
			continue
		}
		if state.Status == StateRUNNING {
			return errors.New("State:" + state.Name + " is " + StateRUNNING + "... Please wait before submitting again")
		}
		if toExecute {
			log.Debug("Execute..." + state.Name)
			errSetRunning := sm.setStateStatusWithTimeStamp(true, state.Name, StateRUNNING, "")
			if errSetRunning != nil {
				log.Debug(errSetRunning.Error())
				return errSetRunning
			}
			state, errSetExecutionID := sm.setExecutionID(state.Name, callerState)
			if errSetExecutionID != nil {
				log.Debug(errSetExecutionID.Error())
				return errSetExecutionID
			}
			err := sm.executeState(*state, callerState, callerOutFile)
			if err != nil {
				errSetFailed := sm.setStateStatusWithTimeStamp(false, state.Name, StateFAILED, "Cmd failed:"+err.Error())
				if errSetFailed != nil {
					return errSetFailed
				}
				return err
			}
			errSetSucceed := sm.setStateStatusWithTimeStamp(false, state.Name, StateSUCCEEDED, "")
			if errSetSucceed != nil {
				return errSetSucceed
			}
		}
		if state.Name == toState {
			break
		}
	}
	return nil
}

//Execute a state
func (sm *States) executeState(state State, callerState *State, callerOutFile *os.File) error {
	log.Debug("Entering... executeState " + state.Name)
	//Check if there is a script
	//Create the log directory if not exists
	dir := filepath.Dir(state.LogPath)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(filepath.Dir(sm.StatesPath), dir)
	}
	errMkDir := os.MkdirAll(dir, 0777)
	if errMkDir != nil {
		logger.AddCallerField().Error(errMkDir.Error())
		return errMkDir
	}
	//Check if log exists and rename it for backup
	outfilePath := filepath.Join(dir, filepath.Base(state.LogPath))
	if _, errLogExists := os.Stat(outfilePath); !os.IsNotExist(errLogExists) {
		newOutfilePath := filepath.Join(dir, filepath.Base(state.LogPath)+"-"+time.Now().Format("2006-01-02T150405.999999-07:00"))
		err := os.Rename(outfilePath, newOutfilePath)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return err
		}
	}
	//Create the log file.
	outfile, err := os.OpenFile(outfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		return err
	}
	defer outfile.Close()
	var errExec error
	// isExtension, errExt := IsExtension(state.Name)
	// if errExt != nil {
	// 	logger.AddCallerField().Error(errExt.Error())
	// 	return errExt
	// }
	if state.IsExtension {
		stateManager, errStateManager := GetStatesManager(state.Name)
		if errStateManager != nil {
			logger.AddCallerField().Error(errStateManager.Error())
			return errStateManager
		}
		errExec = stateManager.Execute(FirstState, LastState, &state, outfile)
	} else {
		if state.Script == "" {
			err := errors.New("The state " + state.Name + " has no script defined")
			logger.AddCallerField().Error(err.Error())
			return err
		}
		log.Debug("script: " + state.Script)
		//Build the command line
		script := state.Script
		if global.Mock {
			timeNow := time.Now().UTC().Format(time.UnixDate)
			script = "echo \"Mock mode: " + timeNow + " extension name: " + sm.ExtensionName + " script for state " + state.Name + " is skipped!\""
		}
		parts := strings.Fields(script)
		var cmd *exec.Cmd
		if len(parts) > 1 {
			cmd = exec.Command(parts[0], parts[1:]...)
		} else {
			cmd = exec.Command(parts[0])
		}
		cmd.Dir = filepath.Dir(sm.StatesPath)
		log.Debug("Execution directory: " + cmd.Dir)
		//Redirect the std to the log file.
		var multiWriter io.Writer
		wOutFile := bufio.NewWriterSize(outfile, 40)
		log.Debug("wOutFile: " + strconv.Itoa(wOutFile.Size()))
		var wCallerOutFile *bufio.Writer
		if callerOutFile != nil {
			wCallerOutFile = bufio.NewWriterSize(callerOutFile, 40)
			multiWriter = io.MultiWriter(wOutFile, wCallerOutFile)
		} else {
			multiWriter = io.MultiWriter(wOutFile)
		}
		cmd.Stdout = multiWriter
		cmd.Stderr = multiWriter
		errExec = cmd.Start()
		if errExec == nil {
			done := make(chan error, 1)
			//Wait signal from channel
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error("Error while running "+state.Name+" ", r)
					}
				}()
				done <- cmd.Wait()
			}()
			select {
			case <-time.After(time.Duration(state.ScriptTimeout) * time.Minute):
				log.Debug("Start Test timeout of " + state.Name)
				if state.ScriptTimeout != 0 {
					if err := cmd.Process.Kill(); err != nil {
						log.Fatal("failed to kill: ", err)
					}
					errExec = errors.New("State " + state.Name + " killed as timeout reached")
				}
				log.Debug("End Test timeout of " + state.Name)
			case err := <-done:
				log.Debug("End of processing of " + state.Name)
				if err != nil {
					errExec = errors.New("process done with error = " + err.Error())
				}
			}
		}
		if callerOutFile != nil {
			logger.AddCallerField().Debug("wCallerOutFile.Flush()")
			err := wCallerOutFile.Flush()
			if err != nil {
				logger.AddCallerField().Error(err.Error())
			}
			logger.AddCallerField().Debug("callerOutFile.Sync()")
			err = callerOutFile.Sync()
			if err != nil {
				logger.AddCallerField().Error(err.Error())
			}
			// callerOutFile.Close()
		}
		logger.AddCallerField().Debug("wOutFile.Flush()")
		err := wOutFile.Flush()
		if err != nil {
			logger.AddCallerField().Error(err.Error())
		}
	}
	logger.AddCallerField().Debug("outfile.Sync()")
	err = outfile.Sync()
	if err != nil {
		logger.AddCallerField().Error(err.Error())
	}
	logger.AddCallerField().Debug("outfile.Close()")
	err = outfile.Close()
	if err != nil {
		logger.AddCallerField().Error(err.Error())
	}
	if errExec != nil {
		logger.AddCallerField().Error(errExec.Error())
		f, err := os.OpenFile(outfilePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return errExec
		}

		defer f.Close()

		scanner := bufio.NewScanner(f)
		line := ""
		for scanner.Scan() {
			line = scanner.Text()
		}
		log.Debug("last line in log:" + line)
		if line != "" {
			errExec = errors.New(errExec.Error() + "\n" + line)
		}
		log.Debug("errExec:" + errExec.Error())
		if _, err = f.WriteString("\nstate:" + state.Name + "\nscript:" + state.Script + "\nlog:" + state.LogPath + "\n" + errExec.Error()); err != nil {
			logger.AddCallerField().Error(err.Error())
		}
	}
	return errExec
}
