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

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extension"
	"gonum.org/v1/gonum/graph"

	log "github.com/sirupsen/logrus"

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
	//StatesToRerun List of states to rerun after this one get executed.
	StatesToRerun []string `yaml:"states_to_rerun" json:"states_to_rerun"`
	//PreviousStates List of previous states, this is not taken into account for the topology sort.
	PreviousStates []string `yaml:"previous_states" json:"previous_states"`
	//NextStates List of next states, this determines the order in which the states will get executed. A topological sort is used to determine the order.
	//If this attribute is not defined in any state of the states file then the NextStates for each state will be set by default to the next state in the states file.
	NextStates []string `yaml:"next_states" json:"next_states"`
}

type States struct {
	StateArray []State `yaml:"states" json:"states"`
	StatesPath string  `yaml:"-" json:"-"`
	mux        *sync.Mutex
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
		mux:        &sync.Mutex{},
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
	// sm.lock()
	//	defer sm.unlock()
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
			dir := extension.GetExtensionLogsPathEmbedded()
			if sm.isCustomStatePath() {
				dir = extension.GetExtensionLogsPathCustom()
			}
			sm.StateArray[index].LogPath = filepath.Join(dir, sm.StateArray[index].Name, sm.StateArray[index].Name+".log")
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
	//if not migrated then set the nextStates
	if !isNextStateMigrationDone {
		sm.setNextStates()
		sm.copyStateToRerunToNextStates()
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
				if !sm.isInNextState(sm.StateArray[index], stateRerun) {
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
func (sm *States) GetStates(status string, extensionsOnly bool, recursive bool) (*States, error) {
	log.Debug("Entering... GetStates")
	errStates := sm.readStates()
	if errStates != nil {
		return nil, errStates
	}
	var states States
	for i := 0; i < len(sm.StateArray); i++ {
		if strings.HasPrefix(sm.StateArray[i].Script, "cm extension") {
			states.StateArray = append(states.StateArray, sm.StateArray[i])
			if recursive {
				smp, err := getStateManager(sm.StateArray[i].Name)
				if err != nil {
					return nil, err
				}
				subStates, err := smp.GetStates(status, extensionsOnly, recursive)
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
		resultStates = states
	}
	return &resultStates, nil
}

//SetStates Set the current states with a new states. If overwrite is false, then the 2 states will be merged.
//States marked deleted in the new states will be removed for the current states.
func (sm *States) SetStates(states States, overwrite bool) error {
	log.Debug("Entering... SetStates")
	sm.lock()
	defer sm.unlock()
	states.setDefaultValues()
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

//indexState Search index in states array for a given state
func indexState(states []State, stateName string) (int, error) {
	for index, stateAux := range states {
		if stateAux.Name == stateName {
			return index, nil
		}
	}
	return -1, errors.New(stateName + " not found")
}

//getStatePosition Search index in the current statearray
func (sm *States) getStatePosition(stateName string) (int, error) {
	index, err := indexState(sm.StateArray, stateName)
	if err != nil {
		return -1, err
	}
	return index, nil
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
			oldIndex, _ := indexState(sm.StateArray, state.Name)
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
			newIndex, _ := indexState(newStates.StateArray, state.Name)
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
	cleanedStates, err := sm.removeDeletedStates(newStates)
	if err != nil {
		return err
	}
	//If no state are defined use the new provided stateArray, no merge needed.
	if len(sm.StateArray) == 0 {
		sm.StateArray = cleanedStates.StateArray
		return nil
	}
	log.Debug("Topology sort")
	newGraph, statesNodesID, statesMap := cleanedStates.addNodes()
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
	newGraph, err = cleanedStates.addEdges(newGraph, statesNodesID)
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
func (sm *States) isInNextState(currentState State, stateName string) bool {
	for _, nextStateName := range currentState.NextStates {
		if nextStateName == stateName {
			return true
		}
	}
	return false
}

//isInPreviousState Check if a state is in the PreviousState of a given state.
func (sm *States) isInPreviousState(currentState State, stateName string) bool {
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
func (sm *States) isRunning() bool {
	for i := 0; i < len(sm.StateArray); i++ {
		state := sm.StateArray[i]
		if state.Status == StateRUNNING {
			return true
		}
	}
	return false
}

//setStateStatus Set the status of a given states. I
//f recusively is true and if the state is an extension then the states of the extension will be set to the status and this recursively.
func (sm *States) setStateStatus(state State, status string, recursively bool) error {
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
		isExtension, err := extension.IsExtension(state.Name)
		if err != nil {
			return err
		}
		if isExtension {
			extensionPath, err := extension.GetRegisteredExtensionPath(state.Name)
			if err != nil {
				return err
			}
			extensionStateManager, err := NewStateManager(extensionPath + string(filepath.Separator) + "statesFile-" + state.Name + ".yml")
			if err != nil {
				return err
			}
			for _, state := range extensionStateManager.StateArray {
				err := sm.setStateStatus(state, status, true)
				if err != nil {
					return err
				}
			}
		}
	}
	errStates := sm.writeStates()
	return errStates
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

//GetState return a state providing its name
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

//setDependencyStatus Set the status and reason for each state referencing the currentState.
func (sm *States) setDependencyStatus(isStart bool, currentState string, status string, reason string) error {
	log.Debugf("Entering in... setDependencyStatus")
	for _, state := range sm.StateArray {
		log.Debug("Check dependency for state:" + state.Name)
		for _, rerunStateName := range state.StatesToRerun {
			if rerunStateName == currentState {
				log.Debug("Dependency " + rerunStateName + " found")
				log.Debug("Current dependency state: " + state.Name + " status " + state.Status)
				if state.Status != StateSKIP {
					log.Debug("set to status " + status + " State: " + rerunStateName)
					err := sm.setStateStatusWithTimeStamp(isStart, state.Name, status, reason)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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

//InsertState Insert state at a given position, before or after a given state.
//If the position is 0 and the stateName is not provided then the state will be inserted taking into account the PreviousStates and NextStates of the inserted state.
//Array start in Go at 0 but here the pos 1 is the elem 0
func (sm *States) InsertState(state State, pos int, stateName string, before bool) error {
	log.Debug("Entering..... InsertState")
	log.Debug("State name: " + stateName)
	sm.lock()
	defer sm.unlock()
	errStates := sm.readStates()
	if errStates != nil {
		return errStates
	}
	_, err := sm._getState(state.Name)
	if err == nil {
		return errors.New("State name " + state.Name + " already exists")
	}
	valid, err := extension.IsExtension(state.Name)
	if err != nil {
		return nil
	}
	if !valid {
		err = errors.New("The state name " + state.Name + " is not a valid extension")
		log.Debug(err.Error())
		return errors.New(err.Error())
	}
	registered := extension.IsExtensionRegistered(state.Name)
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
	} else if position != 0 && (position < 1 || position > len(sm.StateArray)) {
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
		//Update the PreviousStates and NextStates surrounding states
		if arrayPos > 0 {
			if !sm.isInNextState(sm.StateArray[arrayPos-1], state.Name) {
				sm.StateArray[arrayPos-1].NextStates = append(sm.StateArray[arrayPos-1].NextStates, state.Name)
			}
			// if !sm.isInPreviousState(state, sm.StateArray[arrayPos-1].Name) {
			// 	state.PreviousStates = append(state.PreviousStates, sm.StateArray[arrayPos-1].Name)
			// }
		}
		if arrayPos < len(sm.StateArray) {
			// if !sm.isInPreviousState(sm.StateArray[arrayPos], state.Name) {
			// 	sm.StateArray[arrayPos].PreviousStates = append(sm.StateArray[arrayPos].PreviousStates, state.Name)
			// }
			if !sm.isInNextState(state, sm.StateArray[arrayPos].Name) {
				state.NextStates = append(state.NextStates, sm.StateArray[arrayPos].Name)
			}
		}
	} else {
		//Update the NextState of the PreviousStates
		for _, stateName := range state.PreviousStates {
			statePos, err := sm.getStatePosition(stateName)
			if err != nil {
				return err
			}
			if !sm.isInNextState(sm.StateArray[statePos], stateName) {
				sm.StateArray[statePos].NextStates = append(sm.StateArray[statePos].NextStates, state.Name)
			}
		}
	}
	log.Debug(strconv.Itoa(arrayPos))
	sm.StateArray = append(sm.StateArray, state)
	err = sm.topoSort() //	err = sm.hasCycles()
	if err != nil {
		log.Debugf("bckStateArray: %v", bckStateArray)
		sm.StateArray = make([]State, 0)
		sm.StateArray = append(sm.StateArray, bckStateArray...)
		return errors.New(err.Error())
	} else {
		return sm.writeStates()
	}
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
	err = sm.topoSort()
	if err != nil {
		return err
	}
	return sm.writeStates()
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
	states, err := sm.GetStates("", false, false)
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
			logPath = commandsRunner.LogPath
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

//Start states from beginning to end
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
