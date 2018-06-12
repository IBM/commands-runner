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
package stateManager

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

	"gonum.org/v1/gonum/graph"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/crManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	"github.com/go-yaml/yaml"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
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

const StatesFileErrorMessagePattern = "STATES_FILE_ERROR_MESSAGE:"

const PhaseAtEachRun = "AtEachRun"

type State struct {
	Name          string   `yaml:"name" json:"name"`
	Phase         string   `yaml:"phase" json:"phase"`
	Label         string   `yaml:"label" json:"label"`
	LogPath       string   `yaml:"log_path" json:"log_path"`
	Status        string   `yaml:"status" json:"status"`
	StartTime     string   `yaml:"start_time" json:"start_time"`
	EndTime       string   `yaml:"end_time" json:"end_time"`
	Reason        string   `yaml:"reason" json:"reason"`
	Script        string   `yaml:"script" json:"script"`
	ScriptTimeout int      `yaml:"script_timeout" json:"script_timeout"`
	Protected     bool     `yaml:"protected" json:"protected"`
	Deleted       bool     `yaml:"deleted" json:"deleted"`
	StatesToRerun []string `yaml:"states_to_rerun" json:"states_to_rerun"`
	//PreviousStates is not taken into account for the topology sort.
	PreviousStates []string `yaml:"previous_states" json:"previous_states"`
	NextStates     []string `yaml:"next_states" json:"next_states"`
}

type States struct {
	StateArray []State `yaml:"states" json:"states"`
	StatesPath string  `yaml:"-" json:"-"`
	mux        sync.Mutex
}

var pcmLogTempFile *os.File

//NewClient creates a new client
func NewStateManager(statesPath string) (*States, error) {
	log.Debug("Entering... NewStateManager")
	log.Debug("statesPath :" + statesPath)
	if statesPath == "" {
		return nil, errors.New("statesPath not set")
	}
	//Set the default values
	states := &States{
		StateArray: make([]State, 0),
		StatesPath: statesPath,
	}
	return states, nil
}

func (sm *States) isCustomStatePath() bool {
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
	// Read state file.
	sm.lock()
	defer sm.unlock()
	log.Debugf("sm address %p:", &sm)
	statesData, err := ioutil.ReadFile(sm.StatesPath)
	//	log.Debugf("StatesData=%s", statesData)
	if err != nil {
		return err
	}
	log.Debug("states has been read")
	log.Debug("States:\n" + string(statesData))
	// Parse state file into the States structure
	err = yaml.Unmarshal(statesData, &sm)
	log.Debugf("sm address %p:", &sm)
	if err != nil {
		return err
	}
	sm.setDefaultValues()
	log.Debug("Exiting... readStates")
	return nil
}

//Set default value for states
func (sm *States) setDefaultValues() {
	log.Debug("Entering... setDefaultValues")
	isNextStateMigrationDone := false
	for index := range sm.StateArray {
		//		log.Debug("Check state:" + sm.StateArray[index].Name)
		//		log.Debug("Check Label")
		if sm.StateArray[index].Label == "" {
			sm.StateArray[index].Label = sm.StateArray[index].Name
			//			log.Debug("Set state.Label to " + sm.StateArray[index].Label)
		}
		//		log.Debug("Check status")
		if sm.StateArray[index].Status == "" {
			sm.StateArray[index].Status = StateREADY
		}
		//		log.Debug("Check LogPath/Script")

		if sm.StateArray[index].LogPath == "" {
			dir := extensionManager.GetExtensionLogsPathEmbedded()
			if sm.isCustomStatePath() {
				dir = extensionManager.GetExtensionLogsPathCustom()
			}
			sm.StateArray[index].LogPath = dir + sm.StateArray[index].Name + string(filepath.Separator) + sm.StateArray[index].Name + ".log"
			//			log.Debug("Set state.LogPath to " + sm.StateArray[index].LogPath)
		}
		if sm.StateArray[index].Script == "" {
			//			log.Debug("Set state.Script")
			sm.StateArray[index].Script = "cm extension -e " + sm.StateArray[index].Name + " deploy -w"
			//			log.Debug("Set state.Script to " + sm.StateArray[index].Script)
		}
		//		log.Debug("Check ScriptTimeout")
		if sm.StateArray[index].ScriptTimeout == 0 {
			//			log.Debug("Set state.ScriptTimeout")
			sm.StateArray[index].ScriptTimeout = 60
			//			log.Debug("Set state.ScriptTimeout to " + strconv.Itoa(sm.StateArray[index].ScriptTimeout))
		}
		// Remove state in the nextStates which doesn't exist in the state file.
		for indexNext, nextState := range sm.StateArray[index].NextStates {
			indexState, _ := indexState(sm.StateArray, nextState)
			if indexState == -1 {
				sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates[:indexNext], sm.StateArray[index].NextStates[indexNext+1:]...)
			}
		}
	}
	//if not migrated and the nextStates
	if !isNextStateMigrationDone {
		for index := range sm.StateArray {
			if index < len(sm.StateArray)-1 {
				stateNext := sm.StateArray[index+1]
				if len(sm.StateArray[index].NextStates) == 0 {
					sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates, stateNext.Name)
				}
			}
		}
	}
	log.Debug("Exiting... setDefaultValues")
}

func (sm *States) cleanPreviousNextStates() {
	for index := range sm.StateArray {
		// Remove state in the nextStates which doesn't exist in the state file.
		for indexNext, nextState := range sm.StateArray[index].NextStates {
			indexState, _ := indexState(sm.StateArray, nextState)
			if indexState == -1 {
				sm.StateArray[index].NextStates = append(sm.StateArray[index].NextStates[:indexNext], sm.StateArray[index].NextStates[indexNext+1:]...)
			}
		}
		//Remove state in the previousStates which doesn't exist in the state file.
		for indexPrevious, previousState := range sm.StateArray[index].PreviousStates {
			indexState, _ := indexState(sm.StateArray, previousState)
			if indexState == -1 {
				sm.StateArray[index].PreviousStates = append(sm.StateArray[index].PreviousStates[:indexPrevious], sm.StateArray[index].PreviousStates[indexPrevious+1:]...)
			}

		}
	}
}

//Write states
// We can not use defer on sm.unlock because it is not an argument of the function
func (sm *States) writeStates() error {
	log.Debug("Entering... writeStates")
	log.Debug("Marshal states")
	//	log.Debug(sm )
	sm.lock()
	defer sm.unlock()
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

func (sm *States) convert2ByteArray() ([]byte, error) {
	statesData, err := yaml.Marshal(sm)
	log.Debugf("statesData: %s", string(statesData))
	if err != nil {
		return nil, err
	}
	return statesData, err
}

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

func (sm *States) GetStates(status string) (*States, error) {
	log.Debug("Entering... GetStates")
	errStates := sm.readStates()
	if errStates != nil {
		return nil, errStates
	}
	var states States
	//Filter on status
	if status != "" {
		for i := 0; i < len(sm.StateArray); i++ {
			if sm.StateArray[i].Status == status {
				states.StateArray = append(states.StateArray, sm.StateArray[i])
			}
		}
	} else {
		for i := 0; i < len(sm.StateArray); i++ {
			states.StateArray = append(states.StateArray, sm.StateArray[i])
		}

	}
	return &states, nil
}

//Set a states
func (sm *States) SetStates(states States, overwrite bool) error {
	log.Debug("Entering... SetStates")
	states.setDefaultValues()
	//	states.cleanPreviousNextStates()
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

//Will set the
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
		fromIndex, err = sm.getStatePosition(fromStateName)
		log.Debug("From Position: " + fromStateName + " index: " + strconv.Itoa(fromIndex))
		if err != nil {
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
		toIndex, err = sm.getStatePosition(toStateName)
		log.Debug("To Position: " + toStateName + " index: " + strconv.Itoa(toIndex))
		if err != nil {
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

//Search index in states array for a given state
func indexState(states []State, stateName string) (int, error) {
	for index, stateAux := range states {
		if stateAux.Name == stateName {
			return index, nil
		}
	}
	return -1, errors.New(stateName + "not found")
}

//Search index in the current statearray
func (sm *States) getStatePosition(stateName string) (int, error) {
	index, err := indexState(sm.StateArray, stateName)
	if err != nil {
		return -1, err
	}
	return index, nil
}

//Remove deleted states from the current state and new states.
func (sm *States) removeDeletedStates(newStates States) (States, error) {
	log.Debug("Remove states with delete true")
	//   	copyStates := make([]State, len(sm.StateArray))
	copyNewStates := make([]State, len(newStates.StateArray))
	//Remove the states marked for deletion
	//copy(copyStates, sm.StateArray)
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
			oldIndex, _ := indexState(sm.StateArray, state.Name)
			log.Debug("currentStates located at " + strconv.Itoa(oldIndex))
			//if found
			if oldIndex != -1 {
				//if last one then takes the first elements otherwise remove the one in the middle
				if oldIndex == len(sm.StateArray)-1 {
					sm.StateArray = sm.StateArray[:oldIndex]
				} else {
					sm.StateArray = append(sm.StateArray[:oldIndex], sm.StateArray[oldIndex+1:]...)
				}
			}
			newIndex, _ := indexState(newStates.StateArray, state.Name)
			log.Debug("newStates located at " + strconv.Itoa(newIndex))
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
	//	sm.cleanPreviousNextStates()
	err := sm.topoSort()
	if err != nil {
		return newStates, err
	}
	err = newStates.topoSort()
	if err != nil {
		return newStates, err
	}
	return newStates, nil
}

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

func (sm *States) addEdges(newGraph *simple.DirectedGraph, statesNodesID map[string]int64) (*simple.DirectedGraph, error) {
	//Add edges based on state sequence in states file.
	for _, state := range sm.StateArray {
		err := sm.addEdgesNext(state, newGraph, statesNodesID)
		if err != nil {
			return nil, err
		}
	}
	//Add edges based on states_to_rerun
	for _, state := range sm.StateArray {
		ns := newGraph.Node(statesNodesID[state.Name])
		for _, stateToRerun := range state.StatesToRerun {
			if id, ok := statesNodesID[stateToRerun]; ok {
				ne := newGraph.Node(id)
				e := newGraph.NewEdge(ns, ne)
				newGraph.SetEdge(e)
				log.Debug("Add Egde from states_to_rerun: " + state.Name + " -> " + stateToRerun)
			} else {
				log.Warning("WARNING: The state to rerun " + stateToRerun + " listed in states_to_rerun attribute of " + state.Name + " does not exist")
			}
		}
	}
	return newGraph, nil
}

//Generate directed graph
func (sm *States) generateStatesGraph() (*simple.DirectedGraph, map[int64]*State, error) {
	newGraph, statesNodesID, statesMap := sm.addNodes()
	newGraph, err := sm.addEdges(newGraph, statesNodesID)
	return newGraph, statesMap, err
}

func (sm *States) topoSort() error {
	log.Debug("Entering in... topoSort")
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

func (sm *States) topoSortGraph(graph *simple.DirectedGraph, statesMap map[int64]*State) error {
	sorted, err := topo.Sort(graph)
	//Error means cycles
	if err != nil {
		cycles := searchCycleOnGraph(graph, statesMap)
		return generateCyclesError(cycles)
	}
	//Generate the new states based on the sort
	sm.generateStatesFromGraph(sorted, graph, statesMap)
	return nil
}

func (sm *States) searchCycles() ([]*States, error) {
	newGraph, statesMap, err := sm.generateStatesGraph()
	if err != nil {
		return nil, err
	}
	statesCycles := searchCycleOnGraph(newGraph, statesMap)
	return statesCycles, nil
}

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

func (sm *States) hasCycles() error {
	cycles, err := sm.searchCycles()
	if err != nil {
		return err
	}
	log.Debugf("sm.StateArray: %v", sm.StateArray)
	return generateCyclesError(cycles)
}

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

//Merge 2 states
func (sm *States) mergeStates(newStates States) error {
	log.Debug("Entering.... mergeStates")
	newStates, err := sm.removeDeletedStates(newStates)
	if err != nil {
		return err
	}
	//If no state are defined use the new provided stateArray, no merge needed.
	if len(sm.StateArray) == 0 {
		sm.StateArray = newStates.StateArray
		return nil
	}
	log.Debug("Topology sort")
	newGraph, statesNodesID, statesMap := newStates.addNodes()
	//Create and add nodes for the old States if node not yet in the Graph
	for i := 0; i < len(sm.StateArray); i++ {
		//if already inserted then update the state with the current status and other values
		//otherwize insert it as a new node.
		if _, ok := statesNodesID[sm.StateArray[i].Name]; ok {
			log.Debug("Update new Node with old status: " + sm.StateArray[i].Name)
			state := statesMap[statesNodesID[sm.StateArray[i].Name]]
			state.Status = sm.StateArray[i].Status
			state.StartTime = sm.StateArray[i].StartTime
			state.EndTime = sm.StateArray[i].EndTime
			state.Reason = sm.StateArray[i].Reason
			log.Debug("NEw State Node Updated with old status: " + state.Name)
		} else {
			log.Debug("Add old Node: " + sm.StateArray[i].Name)
			n := newGraph.NewNode()
			newGraph.AddNode(n)
			statesNodesID[sm.StateArray[i].Name] = n.ID()
			statesMap[n.ID()] = &sm.StateArray[i]
			log.Debug("Old Node added: " + sm.StateArray[i].Name + " with id: " + strconv.FormatInt(n.ID(), 10))
		}
	}

	//Add the new states edges
	newGraph, err = newStates.addEdges(newGraph, statesNodesID)
	if err != nil {
		return err
	}
	//Add the old states edges
	newGraph, err = sm.addEdges(newGraph, statesNodesID)
	if err != nil {
		return err
	}

	//Print all edges
	for _, edge := range newGraph.Edges() {
		ns := edge.From().ID()
		ne := edge.To().ID()
		log.Debug("Edge: " + strconv.FormatInt(ns, 10) + " -> " + strconv.FormatInt(ne, 10))
		log.Debug("Edge: " + statesMap[ns].Name + " -> " + statesMap[ne].Name)
	}

	//Do the topo sort
	err = sm.topoSortGraph(newGraph, statesMap)
	//Error means cycles
	if err != nil {
		return err
	}
	return nil
}

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

func (sm *States) isInNextState(currentState State, stateName string) bool {
	for _, nextStateName := range currentState.NextStates {
		if nextStateName == stateName {
			return true
		}
	}
	return false
}

func (sm *States) isInPreviousState(currentState State, stateName string) bool {
	for _, previousStateNAme := range currentState.PreviousStates {
		if previousStateNAme == stateName {
			return true
		}
	}
	return false
}

//Check if states is running
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

//Check if a status is Running
func (sm *States) isRunning() bool {
	for i := 0; i < len(sm.StateArray); i++ {
		state := sm.StateArray[i]
		if state.Status == StateRUNNING {
			return true
		}
	}
	return false
}

func (sm *States) setStateStatus(state State, status string, recursively bool) error {
	//Test
	// errStates := sm.readStates()
	// if errStates != nil {
	// 	return errStates
	// }
	index, _ := indexState(sm.StateArray, state.Name)
	if index == -1 {
		return errors.New("State: " + state.Name + " not found!")
	}
	if state.Status != StateSKIP {
		log.Debugln("Change status of " + state.Name)
		sm.StateArray[index].Status = status
	}
	sm.StateArray[index].StartTime = ""
	sm.StateArray[index].EndTime = ""
	sm.StateArray[index].Reason = ""
	if recursively {
		isExtension, err := extensionManager.IsExtension(state.Name)
		if err != nil {
			return err
		}
		if isExtension {
			extensionPath, err := extensionManager.GetRegisteredExtensionPath(state.Name)
			if err != nil {
				return err
			}
			extensionStateManager, err := NewStateManager(extensionPath + string(filepath.Separator) + "statesFile-" + state.Name + ".yml")
			if err != nil {
				return err
			}
			err = extensionStateManager.ResetEngine()
			if err != nil {
				return err
			}
		}
	}
	errStates := sm.writeStates()
	return errStates
}

//Reset states, all non-skip state will be set to READY recursively
//No RUNNING state must be found.
func (sm *States) ResetEngine() error {
	log.Debug("Entering... ResetEngine")
	//Read states
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
	//Reset states
	for _, state := range sm.StateArray {
		err := sm.setStateStatus(state, StateREADY, true)
		if err != nil {
			return err
		}
	}
	//Write states
	errStates = sm.writeStates()
	return errStates
}

//Get a state
func (sm *States) GetState(state string) (*State, error) {
	log.Debug("Entering... GetState")
	log.Debugf("Read states=%s\n", state)
	errStates := sm.readStates()
	if errStates != nil {
		return nil, errStates
	}
	stateFound, errState := sm._getState(state)
	return stateFound, errState
}

//Set a state status
func (sm *States) SetState(state string, status string, reason string, script string, scriptTimout int, recursivelly bool) error {
	log.Debugf("Read states=%s\n", state)
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

/*
//Set a state status
func (sm *States) setStateStatus(state string, status string) error {
	log.Debugf("Read states=%s\n", state)
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	stateFound, errState := sm._getState(state)
	if errState != nil {
		return errState
	}
	log.Debugf("Set state %s with status %s", stateFound.Name, status)
	stateFound.Status = status
	stateFound.Reason = ""
	stateFound.StartTime = ""
	stateFound.EndTime = ""
	errWriteStates := sm.writeStates()
	if errWriteStates != nil {
		return errWriteStates
	}
	return nil
}
*/

//Set dependency status
func (sm *States) setDependencyStatus(isStart bool, currentState string, status string, reason string) error {
	log.Debugf("Entering in... setDependencyStatus")
	for _, state := range sm.StateArray {
		log.Debug("Check dependency for state:" + state.Name)
		for _, rerunStateName := range state.StatesToRerun {
			if rerunStateName == currentState {
				log.Debug("Dependency " + rerunStateName + " found")
				log.Debug("Current dependency state: " + state.Name + " status " + state.Status)
				if state.Status == StateSUCCEEDED {
					log.Debug("set to FAILED State: " + rerunStateName)
					err := sm.setStateStatusWithTimeStamp(isStart, state.Name, StateFAILED, reason)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

//Set a state status with timestamp
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
		err := sm.setDependencyStatus(isStart, state, status, "being a dependency of the failed state "+state)
		if err != nil {
			return err
		}
	} else {
		stateFound.Reason = ""
	}
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

//Insert state at a given position
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) InsertState(state State, pos int, stateName string, before bool) error {
	log.Debug("Entering..... InsertState")
	log.Debug("State name: " + stateName)
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	_, err := sm._getState(state.Name)
	if err == nil {
		return errors.New("State name " + state.Name + " already exists")
	}
	valid, err := extensionManager.IsExtension(state.Name)
	if err != nil {
		return nil
	}
	if !valid {
		err = errors.New("The state name " + state.Name + " is not a valid extension")
		log.Debug(err.Error())
		return errors.New(err.Error())
	}
	registered := extensionManager.IsExtensionRegistered(state.Name)
	if !registered {
		return errors.New("The state name " + state.Name + " is not registered")
	}
	if sm.isRunning() {
		return errors.New("Insert can not be executed while a deployment is running")
	}
	position := pos
	if stateName != "" {
		position, err = sm.getStatePosition(stateName)
		if err != nil {
			return err
		}
		position++
		log.Debug("Position:" + strconv.Itoa(position))
	} else if position == 0 && (len(state.NextStates) == 0 || len(state.PreviousStates) == 0) {
		return errors.New("The position, state name and previous and next states are undefined")
	} else if position < 1 || position > len(sm.StateArray) {
		return errors.New("The position must be between 1 and " + strconv.Itoa(len(sm.StateArray)) + " currently:" + strconv.Itoa(position))
	}

	//Copy the state at the end but it will be overwritten by the copy
	bckStateArray := make([]State, 0)
	bckStateArray = append(bckStateArray, sm.StateArray...)
	log.Debugf("%v", bckStateArray)
	arrayPos := position
	if position != 0 {
		if before {
			arrayPos = position - 1
		}
		if arrayPos > 0 {
			if !sm.isInNextState(sm.StateArray[arrayPos-1], state.Name) {
				sm.StateArray[arrayPos-1].NextStates = append(sm.StateArray[arrayPos-1].NextStates, state.Name)
			}
			if !sm.isInPreviousState(state, sm.StateArray[arrayPos-1].Name) {
				state.PreviousStates = append(state.PreviousStates, sm.StateArray[arrayPos-1].Name)
			}
		}
		if arrayPos < len(sm.StateArray) {
			if !sm.isInPreviousState(sm.StateArray[arrayPos], state.Name) {
				sm.StateArray[arrayPos].PreviousStates = append(sm.StateArray[arrayPos].PreviousStates, state.Name)
			}
			if !sm.isInNextState(state, sm.StateArray[arrayPos].Name) {
				state.NextStates = append(state.NextStates, sm.StateArray[arrayPos].Name)
			}
		}
	}
	log.Debug(strconv.Itoa(arrayPos))
	sm.StateArray = append(sm.StateArray, state)
	//	copy(sm.StateArray[arrayPos+1:], sm.StateArray[arrayPos:])
	//	sm.StateArray[arrayPos] = state
	err = sm.topoSort()
	//	err = sm.hasCycles()
	if err != nil {
		log.Debugf("bckStateArray: %v", bckStateArray)
		sm.StateArray = make([]State, 0)
		sm.StateArray = append(sm.StateArray, bckStateArray...)
		return errors.New(err.Error())
	} else {
		return sm.writeStates()
	}
}

//Delete a state at a given position
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) DeleteState(pos int, stateName string) error {
	log.Debug("Entering..... InsertState")
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
		position, err = sm.getStatePosition(stateName)
		if err != nil {
			return err
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
	copy(sm.StateArray[arrayPos:], sm.StateArray[arrayPos+1:])
	sm.StateArray[len(sm.StateArray)-1] = State{} // or the zero value of T
	sm.StateArray = sm.StateArray[:len(sm.StateArray)-1]
	//	sm.cleanPreviousNextStates()
	err = sm.topoSort()
	if err != nil {
		return err
	}
	return sm.writeStates()
}

//Search logPath for a given state in states structure
func (sm *States) getLogPath(state string) (string, error) {
	stateFound, err := sm._getState(state)
	if err != nil {
		return "", err
	}
	logPath := stateFound.LogPath
	if logPath == "" {
		return logPath, errors.New("No logPath available for " + state)
	}
	//	if !filepath.IsAbs(logPath) {
	//		return logPath, errors.New("The logPath :" + logPath + " for state :" + state + " is not absolute")
	//	}
	log.Debugf("LogPath:%s", logPath)
	return logPath, nil
}

func (sm *States) GetLogs(position int64, length int64, bychar bool) (string, error) {
	var data []byte
	states, err := sm.GetStates("")
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

/*Retrieve log of a given state.
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
	if state == "" {
		return nil, errors.New("State must be defined")
	}
	//Mock log
	switch state {
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
			pcmLogTempFile, err = ioutil.TempFile("/tmp/", "/cfp-commands-runner-log")
			if err != nil {
				return nil, err
			}
			logPath = crManager.LogPath
			logFile, err := os.Open(logPath)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(pcmLogTempFile, logFile)
			if err != nil {
				return nil, err
			}
		}
		logPath = pcmLogTempFile.Name()
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
		logPath = logPathFound
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

//Execute states from beginning to end
func (sm *States) Start() error {
	log.Debug("Enterring... Start")
	return sm.Execute(FirstState, LastState)
}

//Execute states from state 'fromState' to state 'toState'
func (sm *States) Execute(fromState string, toState string) error {
	log.Debug("Enterring... Execute")
	log.Debug("State:" + sm.StatesPath)
	log.Debug("From state:" + fromState)
	log.Debug("To   state:" + toState)
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

	err = sm.executeStates(fromState, toState)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	return nil
}

//Execute states
func (sm *States) executeStates(fromState string, toState string) error {
	toExecute := false || fromState == FirstState
	//Relative statesPath is used in unit-test and so in that case
	//we must not change to another directory.
	if filepath.IsAbs(sm.StatesPath) {
		//Search the home dir and change to that directory
		statesFileDir := filepath.Dir(sm.StatesPath)
		err := os.Chdir(statesFileDir)
		if err != nil {
			return err
		}
	}
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
		script := state.Script
		if toExecute && script != "" {
			log.Debug("Execute..." + state.Name)
			errSetRunning := sm.setStateStatusWithTimeStamp(true, state.Name, StateRUNNING, "")
			if errSetRunning != nil {
				log.Debug(errSetRunning.Error())
				return errSetRunning
			}
			err := sm.executeState(state)
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
func (sm *States) executeState(state State) error {
	//Check if there is a script
	if state.Script == "" {
		return nil
	}
	//Build the command line
	script := state.Script
	parts := strings.Fields(script)
	var cmd *exec.Cmd
	if len(parts) > 1 {
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(parts[0])
	}
	//Create the log directory if not exists
	dir := filepath.Dir(state.LogPath)
	errMkDir := os.MkdirAll(dir, 0777)
	if errMkDir != nil {
		panic(errMkDir)
	}
	//Check if log exists and rename it for backup
	if _, errLogExists := os.Stat(state.LogPath); !os.IsNotExist(errLogExists) {
		os.Rename(state.LogPath, state.LogPath+"-"+time.Now().Format("2006-01-02T150405.999999-07:00"))
	}
	//Create the log file.
	outfile, err := os.Create(state.LogPath)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	//Update states to be rerun
	for _, stateName := range state.StatesToRerun {
		stateToReRun, err := sm._getState(stateName)
		if err != nil {
			log.Debug("WARNING: State to rerun " + stateName + " not found in states_to_rerun attribute of " + state.Name + ":" + err.Error())
		} else {
			if stateToReRun.Status != StateSKIP {
				err := sm.setStateStatus(*stateToReRun, StateREADY, true)
				if err != nil {
					return errors.New("State to rerun " + stateToReRun.Name + " not found in states_to_rerun attribute of " + state.Name + ":" + err.Error())
				}
				log.Debug("Reset to READY state " + stateToReRun.Name)
			}
		}
	}
	//Redirect the std to the log file.
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	errExec := cmd.Start()
	done := make(chan error, 1)
	//Wait signal from channel
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Duration(state.ScriptTimeout) * time.Minute):
		if state.ScriptTimeout != 0 {
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			errExec = errors.New("State " + state.Name + " killed as timeout reached")
		}
	case err := <-done:
		if err != nil {
			errExec = errors.New("process done with error = " + err.Error())
		}
	}

	outfile.Sync()
	outfile.Close()
	if errExec != nil {
		f, err := os.Open(state.LogPath)
		if err != nil {
			log.Debug(err.Error())
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
			log.Debug(err.Error())
		}
	}
	return errExec
}
