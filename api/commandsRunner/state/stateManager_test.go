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
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-yaml/yaml"
)

const COPYRIGHT_TEST string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func init() {
	//	log.SetLevel(log.DebugLevel)
	state1log := "sample-state1 line 1\nsample-state1 line 2\nsample-state1 line 3\nsample-state1 line 4\nsample-state1 line 5"
	state2log := "sample-state2 line 1\nsample-state2 line 2\nsample-state2 line 3\nsample-state2 line 4\nsample-state2 line 5"
	ioutil.WriteFile("/tmp/sample-state1.log", []byte(state1log), 0600)
	errstate1 := os.Chmod("/tmp/sample-state1.log", 0600)
	if errstate1 != nil {
		log.Fatal(errstate1.Error())
	}
	ioutil.WriteFile("/tmp/sample-state2.log", []byte(state2log), 0600)
	errstate2 := os.Chmod("/tmp/sample-state2.log", 0600)
	if errstate2 != nil {
		log.Fatal(errstate2.Error())
	}
}

func TestGetStatesOK(t *testing.T) {
	//	log.SetLevel(log.DebugLevel)
	t.Log("Entering... TestGetStatesOK")
	statesPath := "../../test/resource/states-TestGetStatesOK.yaml"
	SetExtensionPath("../../test/data/extensions/")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	//	sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetStatesOK")
	sm.StatesPath = statesPath
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	states, err := sm.GetStates("", false, false)
	if err != nil {
		t.Error(err.Error())
	}
	stateM, _ := json.Marshal(states)
	var statesIn States
	statesData, err := ioutil.ReadFile(statesPath)
	if err != nil {
		t.Fatal(err)
	}
	// Parse state file into the States structure
	err = yaml.Unmarshal(statesData, &statesIn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("States YAML:%v", statesIn)
	expected, _ := json.Marshal(statesIn)
	got := stateM
	t.Log(string(expected) + "\n")
	t.Log(string(got))
	if string(got) != string(expected) {
		t.Errorf("handler returned unexpected response: got \n\n%v \n\nwant \n\n%v",
			string(stateM), string(expected))
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetStatesWithStatus(t *testing.T) {
	t.Log("Entering... TestGetStatesWithStatus")
	//	sm, err := newStateManager("../../test/resource/states-run-running.yaml")
	statesPath := "../../test/resource/states-run-running.yaml"
	sm := newStateManager("states-run-running")
	sm.StatesPath = statesPath
	states, err := sm.GetStates(StateRUNNING, false, false)
	if err != nil {
		t.Error(err.Error())
	}
	if len(states.StateArray) != 1 {
		for i := 0; i < len(states.StateArray); i++ {
			t.Logf("State:%s", states.StateArray[i].Name)
		}
		t.Error("Expected 1 state got:" + strconv.Itoa(len(states.StateArray)))
	}
}

func TestSetStatesOK(t *testing.T) {
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-post-sample-from-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//	sm, err := newStateManager("../../test/resource/out-test-states-post-sample-from-json.yaml")
	statesPath := "../../test/resource/out-test-states-post-sample-from-json.yaml"
	sm := newStateManager("out-test-states-post-sample-from-json")
	sm.StatesPath = statesPath
	statesJSON := "{\"states\":[{\"name\":\"state1\",\"label\":\"state1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	t.Log("Entering... TestSetStatesOK")
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStatesStatusesOK(t *testing.T) {
	t.Log("Entering... TestSetStatesStatusesOK")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	//	sm, err := newStateManager("../../test/resource/states-TestSetStatesStatusesOK.yaml")
	statesPath := "../../test/resource/states-TestSetStatesStatusesOK.yaml"
	sm := newStateManager("states-TestSetStatesStatusesOK")
	sm.StatesPath = statesPath
	err := sm.SetStatesStatuses("SKIP", "state1", true, "state1", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err := sm.GetState("state1")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status != StateSKIP {
		t.Error("Status not set as SKIP as expected:" + state.Status)
	}
	states, err := sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) > 1 {
		t.Error("Another state was set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "state1", true, "state1", true)
	if err != nil {
		t.Error(err.Error())
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStatesStatusesFromTo(t *testing.T) {
	t.Log("Entering... TestSetStatesStatusesFromTo")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	// sm, err := newStateManager("../../test/resource/states-TestSetStatesStatusesFromTo.yaml")
	statesPath := "../../test/resource/states-TestSetStatesStatusesFromTo.yaml"
	sm := newStateManager("states-TestSetStatesStatusesFromTo")
	sm.StatesPath = statesPath
	//Test a range in the middle inclusive
	err := sm.SetStatesStatuses("SKIP", "repeat", true, "nologpath", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err := sm.GetState("state1")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err := sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 3 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	//Test a range in the middle right-exclusive
	err = sm.SetStatesStatuses("SKIP", "repeat", true, "nologpath", false)
	if err != nil {
		t.Error(err.Error())
	}
	state, err = sm.GetState("state1")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP and got :" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 2 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	//Test a range in the middle exclusive
	err = sm.SetStatesStatuses("SKIP", "repeat", false, "nologpath", false)
	if err != nil {
		t.Error(err.Error())
	}
	state, err = sm.GetState("repeat")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 1 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	//Test range in the middle left-exclusive
	err = sm.SetStatesStatuses("SKIP", "repeat", false, "nologpath", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err = sm.GetState("repeat")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 2 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	//Test range from the middle to end
	err = sm.SetStatesStatuses("SKIP", "repeat", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err = sm.GetState("repeat")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status != StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 4 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}

	//Test range from the start to middle
	err = sm.SetStatesStatuses("SKIP", "", true, "repeat", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err = sm.GetState("repeat")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status != StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP, false, false)
	if len(states.StateArray) != 2 {
		t.Error("Not the correct number of states were set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	//Test wrong range
	err = sm.SetStatesStatuses("SKIP", "repeat", true, "state1", true)
	if err == nil {
		t.Error("Range in wrong order error expected")
	}
	t.Log(err.Error())
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}

}

func TestSetStatesWithDelete(t *testing.T) {
	_, err := os.Create("../../test/resource/out-test-states-delete-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-states-delete-json.yaml")
	statesPath := "../../test/resource/out-test-states-delete-json.yaml"
	sm := newStateManager("out-test-states-delete-json")
	sm.StatesPath = statesPath
	statesJSON := "{\"states\":[{\"name\":\"state1\",\"delete:\":true,\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	t.Log("Entering... TestSetStatesOK")
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("state1", false, false)
	if err != nil {
		t.Error(err.Error())
	}
	states = *statesResult
	if len(states.StateArray) > 0 {
		t.Error("State 1 not removed")
	}
	statesResult, err = sm.GetStates("cr", false, false)
	if err != nil {
		t.Error(err.Error())
	}
	states = *statesResult
	if len(states.StateArray) > 0 {
		t.Error("CR removed")
	}
}

func TestSetStatesMerge(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	t.Log("Entering... TestSetStatesMerge")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/out-test-states-merge-json.yaml"
	// err := os.Remove(statesPath)
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	_, err := os.Create(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-states-merge-json.yaml")
	sm := newStateManager("out-test-states-merge-json.yaml")
	sm.StatesPath = statesPath
	statesJSON := "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	statesData, _ := sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr2\",\"label\":\"commands-runner2\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	//var states States
	json.Unmarshal([]byte(statesJSON), &states)
	t.Log(states)
	err = sm.SetStates(states, false)
	statesData, _ = sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("", false, false)
	if err != nil {
		t.Error(err.Error())
	}
	if statesResult.StateArray[0].Name != "state1" &&
		!(statesResult.StateArray[1].Name == "cr" || statesResult.StateArray[1].Name == "cr2") &&
		!(statesResult.StateArray[2].Name == "cr2" || statesResult.StateArray[2].Name == "cr") {
		t.Error("Wrong order")
	}
}

func TestSetStatesMergeWithDelete(t *testing.T) {
	t.Log("Entering... TestSetStatesMergeWithDelete")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-merge-delete-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-states-merge-delete-json.yaml")
	statesPath := "../../test/resource/out-test-states-merge-delete-json.yaml"
	sm := newStateManager("out-test-states-merge-delete-json")
	sm.StatesPath = statesPath
	statesJSON := "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"SUCCEEDED\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr2\",\"label\":\"commmands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"deleted\":true,\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, false)
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("", false, false)
	if err != nil {
		t.Error(err.Error())
	}
	if len(statesResult.StateArray) != 2 {
		t.Error("Result with size " + strconv.Itoa(len(statesResult.StateArray)) + " and expect 2")
	}
	if statesResult.StateArray[0].Name != "state1" &&
		statesResult.StateArray[1].Name != "cr2" {
		t.Error("Wrong order")
	}
	if statesResult.StateArray[0].Status != StateSUCCEEDED {
		t.Error("State1 doesn't have the correct status. expecting SUCCEEDED got " + statesResult.StateArray[0].Status)
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStatesMergeWithCycle(t *testing.T) {
	t.Log("Entering... TestSetStatesMergeWithCycle")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-merge-cycle-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//	sm, err := newStateManager("../../test/resource/out-test-states-merge-cycle-json.yaml")
	statesPath := "../../test/resource/out-test-states-merge-cycle-json.yaml"
	sm := newStateManager("out-test-states-merge-cycle-json")
	sm.StatesPath = statesPath
	statesJSON := "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"state1\",\"label\":\"State 1\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cr\",\"label\":\"commands-runner\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\",\"next_states\":[\"state1\"]}]}"
	json.Unmarshal([]byte(statesJSON), &states)
	statesData, _ := states.convert2String()
	t.Log(statesData)
	err = sm.SetStates(states, false)
	statesData, _ = sm.convert2String()
	t.Log(statesData)
	if err == nil {
		t.Error("Expecting error as there is cycles")
	} else {
		t.Log(err.Error())
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestInsertStatesBeforeFirst(t *testing.T) {
	t.Log("Entering... TestInsertStatesBeforeFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-before-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//	sm, err := newStateManager("../../test/resource/out-test-insert-before-first-state.yaml")
	statesPath := "../../test/resource/out-test-insert-before-first-state.yaml"
	sm := newStateManager("out-test-insert-before-first-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	state := &State{
		Name: "ext-template",
	}
	err = sm.InsertState(*state, 1, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "ext-template" ||
		sm.StateArray[1].Name != "First" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesAfterFirst(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//	sm, err := newStateManager("../../test/resource/out-test-insert-after-first-state.yaml")
	statesPath := "../../test/resource/out-test-insert-after-first-state.yaml"
	sm := newStateManager("out-test-insert-after-first-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("ext-template")
	state := &State{
		Name: "ext-template",
	}
	err = sm.InsertState(*state, 1, "", false)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "ext-template" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
}
func TestInsertStatesBeforeLast(t *testing.T) {
	t.Log("Entering... TestInsertStatesBeforeLast")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-before-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-insert-before-last-state.yaml")
	statesPath := "../../test/resource/out-test-insert-before-last-state.yaml"
	sm := newStateManager("out-test-insert-before-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	err = sm.SetStates(*states, true)
	stateData, _ := sm.convert2String()
	t.Log(stateData)
	if err != nil {
		t.Error(err.Error())
	}
	state := &State{
		Name: "ext-template",
	}
	err = sm.InsertState(*state, 2, "", true)
	stateData, _ = sm.convert2String()
	t.Log(stateData)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "ext-template" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
	//	t.Error("")
}

func TestInsertStatesAfterLast(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-insert-after-last-state.yaml")
	statesPath := "../../test/resource/out-test-insert-after-last-state.yaml"
	sm := newStateManager("out-test-insert-after-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("ext-template")
	state := &State{
		Name: "ext-template",
	}
	err = sm.InsertState(*state, 2, "", false)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "Last" ||
		sm.StateArray[2].Name != "ext-template" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesAfterLastByName(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-insert-after-last-state.yaml")
	statesPath := "../../test/resource/out-test-insert-after-last-state.yaml"
	sm := newStateManager("out-test-insert-after-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	statesData, _ := sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("ext-template")
	state := &State{
		Name: "ext-template",
	}
	err = sm.InsertState(*state, 0, "Last", false)
	statesData, _ = sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "Last" ||
		sm.StateArray[2].Name != "ext-template" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesWithCycle(t *testing.T) {
	t.Log("Entering... TestInsertStatesWithCycle")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-cycle-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-insert-cycle-state.yaml")
	statesPath := "../../test/resource/out-test-insert-cycle-state.yaml"
	sm := newStateManager("out-test-insert-cycle-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	state := &State{
		Name: "ext-template",
		NextStates: []string{
			"First",
			"Last",
		},
	}
	err = sm.InsertState(*state, 2, "", true)
	stateData, _ := sm.convert2String()
	t.Log(stateData)
	if err == nil {
		t.Error("Expecting error as the state insert generate a cycle")
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
}

func TestDeleteStatesFirst(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-delete-first-state.yaml")
	statesPath := "../../test/resource/out-test-delete-first-state.yaml"
	sm := newStateManager("out-test-delete-first-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.DeleteState(1, "")
	if err != nil {
		t.Error(err.Error())
	}
	if len(sm.StateArray) > 1 ||
		sm.StateArray[0].Name != "Last" {
		t.Error("State not deleted")
	}
}

func TestDeleteStatesLast(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-delete-last-state.yaml")
	statesPath := "../../test/resource/out-test-delete-last-state.yaml"
	sm := newStateManager("out-test-delete-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.DeleteState(2, "")
	if err != nil {
		t.Error(err.Error())
	}
	if len(sm.StateArray) > 1 ||
		sm.StateArray[0].Name != "First" {
		t.Error("State not delted")
	}
}

func TestDeleteStatesFirstByName(t *testing.T) {
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	t.Log("Entering... TestInsertStatesAfterFirst")
	_, err := os.Create("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//	sm, err := newStateManager("../../test/resource/out-test-delete-first-state.yaml")
	statesPath := "../../test/resource/out-test-delete-first-state.yaml"
	sm := newStateManager("out-test-delete-first-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.DeleteState(0, "First")
	if err != nil {
		t.Error(err.Error())
	}
	if len(sm.StateArray) > 1 ||
		sm.StateArray[0].Name != "Last" {
		t.Error("State not delted")
	}
}

func TestDeleteStatesLastByName(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-delete-last-state.yaml")
	statesPath := "../../test/resource/out-test-delete-last-state.yaml"
	sm := newStateManager("out-test-delete-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name: "First",
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.DeleteState(0, "Last")
	if err != nil {
		t.Error(err.Error())
	}
	if len(sm.StateArray) > 1 ||
		sm.StateArray[0].Name != "First" {
		t.Error("State not delted")
	}
}

func TestDeleteStatesProtected(t *testing.T) {
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	t.Log("Entering... TestInsertStatesAfterFirst")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	// sm, err := newStateManager("../../test/resource/out-test-delete-last-state.yaml")
	statesPath := "../../test/resource/out-test-delete-last-state.yaml"
	sm := newStateManager("out-test-delete-last-state")
	sm.StatesPath = statesPath
	states := &States{
		StateArray: []State{
			{
				Name:      "First",
				Protected: true,
			},
			{
				Name: "Last",
			},
		},
	}
	t.Log("Set States")
	err = sm.SetStates(*states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.DeleteState(1, "")
	if err != nil {
		t.Log(err.Error())
	}
	if err == nil {
		t.Error("Expecting error as state is protected")
	}
}
func TestGetStateOK(t *testing.T) {
	t.Log("Entering... TestGetStateOK")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	// sm, err := newStateManager("../../test/resource/states-TestGetStateOK.yaml")
	statesPath := "../../test/resource/states-TestGetStateOK.yaml"
	sm := newStateManager("states-TestGetStateOK")
	sm.StatesPath = statesPath
	state, err := sm.GetState("state1")
	if err != nil {
		t.Error(err.Error())
	}
	stateM, _ := yaml.Marshal(state)
	t.Log("State found:" + string(stateM))
	if state.Status != "READY" {
		t.Error("Expecting 'READY' in reponse but gets:" + state.Status)
	}
}

func TestGetStateNOK(t *testing.T) {
	t.Log("Entering... TestGetStateNOK")
	// sm, err := newStateManager("../../test/resource/states-TestGetStateNOK.yaml")
	statesPath := "../../test/resource/states-TestGetStateNOK.yaml"
	sm := newStateManager("states-TestGetStateNOK")
	sm.StatesPath = statesPath
	stateN, err := sm.GetState("not-exist")
	if err == nil {
		t.Error("Found a state but should'nt " + stateN.Status)
	}
}

func TestGetStateEmptyState(t *testing.T) {
	t.Log("Entering... TestGetStateNOK")
	// sm, err := newStateManager("../../test/resource/states-TestGetStateEmptyState.yaml")
	statesPath := "../../test/resource/states-TestGetStateEmptyState.yaml"
	sm := newStateManager("states-TestGetStateEmptyState")
	sm.StatesPath = statesPath
	_, err := sm.GetState("")
	if err == nil {
		t.Error("An error should be raised as the state is not specified")
	}
}

func TestSetStateStatus(t *testing.T) {
	t.Log("Entering... TestSetStateStatus")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	stateV := StateRUNNING
	reasonV := strconv.FormatInt(time.Now().UnixNano(), 5)
	scriptV := strconv.FormatInt(time.Now().UnixNano(), 7)
	scriptTimeoutV := time.Now().Second()
	statesPath := "../../test/resource/states-TestSetStateStatus.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestSetStateStatus")
	sm.StatesPath = statesPath
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.SetState("cr", stateV, reasonV, scriptV, scriptTimeoutV, true)
	if err != nil {
		t.Error(err.Error())
	}
	stateN, err := sm.GetState("cr")
	if err != nil {
		t.Error(err.Error())
	}
	if stateN.Status != stateV {
		t.Error("Expected:" + stateV + " gets:" + stateN.Status)
	}
	if stateN.Reason != "" {
		t.Error("Expected: \"\" gets:" + stateN.Reason)
	}
	if stateN.Script != scriptV {
		t.Error("Expected:" + scriptV + " gets:" + stateN.Script)
	}
	if stateN.ScriptTimeout != scriptTimeoutV {
		t.Error("Expected:" + strconv.Itoa(scriptTimeoutV) + " gets:" + strconv.Itoa(stateN.ScriptTimeout))
	}
	err = sm.SetState("cr", StateFAILED, reasonV, scriptV, scriptTimeoutV, true)
	if err != nil {
		t.Error(err.Error())
	}
	stateN, err = sm.GetState("cr")
	if err != nil {
		t.Error(err.Error())
	}
	if stateN.Reason != reasonV {
		t.Error("Expected:" + reasonV + " gets:" + stateN.Reason)
	}
	err = sm.SetState("cr", StateFAILED, "", "hello.sh", 61, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.SetState("cr", "WrongStatus", "", "hello.sh", 61, true)
	if err == nil {
		t.Error("Expecting error as the status is incorrect")
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineSuccess(t *testing.T) {
	t.Log("Entering...TestEngineSuccess")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-success.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-success")
	sm.StatesPath = statesPath
	t.Log("Reset States file")
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates(StateFAILED, false, false)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) > 0 {
		t.Error("At least one state failed:" + states.StateArray[0].Name)
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineWithRerunAfter(t *testing.T) {
	t.Log("Entering... TestEngineWithRerunAfter")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-with-rerun-after.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-with-rerun-after")
	sm.StatesPath = statesPath
	t.Log("Reset States file")
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates(StateFAILED, false, false)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) > 0 {
		t.Error("At least one state failed:" + states.StateArray[0].Name)
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineWithRerunBefore(t *testing.T) {
	t.Log("Entering... TestEngineWithRerunAfter")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-with-rerun-before.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-with-rerun-before")
	sm.StatesPath = statesPath
	t.Log("Reset States file")
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	err = sm.Execute("task1", "task3")
	if err == nil {
		t.Error("Expect error because the stateToRerun of task2 reference task1 which is before in the sequence")
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineCycleOnRerun(t *testing.T) {
	t.Log("Entering... TestEngineCycleOnRerun")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-cycle-on-rerun.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-cycle-on-rerun")
	sm.StatesPath = statesPath
	t.Log("Execute states file")
	err := sm.Start()
	if err == nil {
		t.Error("Expected error as it has a cyle task1->task2")
	} else {
		t.Log(err.Error())
	}
}

func TestEngineCycleOnNext(t *testing.T) {
	t.Log("Entering... TestEngineCycleOnNext")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-cycle-on-next.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-cycle-on-next")
	sm.StatesPath = statesPath
	t.Log("Execute states file")
	err := sm.Start()
	if err == nil {
		t.Error("Expected error as it has a cyle task1->task2->task1")
	} else {
		t.Log(err.Error())
	}
}

func TestNextStatusSet(t *testing.T) {
	state1 := State{
		Name: "task1",
	}
	state2 := State{
		Name: "task2",
	}
	stateArray := []State{state1, state2}
	sm := &States{
		StateArray: stateArray,
	}
	sm.topoSort()
	statesData, _ := sm.convert2String()
	t.Log(statesData)
	if len(sm.StateArray[0].NextStates) == 0 {
		t.Error("next_states not updated")
	}
	t.Logf("%+v", sm.StateArray[0].NextStates)
}
func TestStatusFailedDependency(t *testing.T) {
	t.Log("Entering... TestStatusFailedDependency")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-failed-dependency.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-failed-dependency")
	sm.StatesPath = statesPath
	t.Log("Reset States file")
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates("", false, false)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if states.StateArray[0].Status != StateSUCCEEDED ||
		states.StateArray[1].Status != StateFAILED ||
		states.StateArray[2].Status != StateFAILED {
		t.Error("The statuses are not correct, expecting task1: SUCCEEDED, task2: FAILED, task3: FAILED")
	}
	t.Log(states)
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineFailure(t *testing.T) {
	t.Log("Entering... TestEngineFailure")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-failure.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-run-failure")
	sm.StatesPath = statesPath
	t.Log("Reset States file")
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	err = sm.Execute("task1", "task3")
	if err == nil {
		t.Error("Expecting error as the execution should fail")
	}
	time.Sleep(10 * time.Second)
	t.Log("Get Failed states")
	states, errStates := sm.GetStates(StateFAILED, false, false)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) <= 0 {
		t.Error("At least one state must failed")
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStateStatusEmptyState(t *testing.T) {
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-TestSetStateStatusEmptyState.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestSetStateStatusEmptyState")
	sm.StatesPath = statesPath
	err := sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Entering... TestSetStateStatus")
	err = sm.SetState("", "TEST", "REASON", "", -1, true)
	if err == nil {
		t.Error("An error should be raised as the state is not specified")
	}
	sm.ResetEngineExecutionInfo()
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetLogGoodState1(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	t.Log("Entering... TestGetLogGoodState1")
	statesPath := "../../test/resource/states-TestGetLogGoodState1.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodState1")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "state1") {
		t.Error("Expecting 'state1' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodState1ByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodState1ByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodState1ByChar.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodState1ByChar")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 0, math.MaxInt64, true)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "state1") {
		t.Error("Expecting 'state1' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodCR(t *testing.T) {
	t.Log("Entering... TestGetLogGoodCR")
	//log.SetLevel(log.DebugLevel)
	statesPath := "../../test/resource/states-TestGetLogGoodState2.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodState2")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state2", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "state2") {
		t.Error("Expecting 'state2' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodMock(t *testing.T) {
	t.Log("Entering... TestGetLogGoodMock")
	statesPath := "../../test/resource/states-TestGetLogGoodMock.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodMock")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("mock", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "Mock") {
		t.Error("Expecting 'Mock' in reponse but gets:" + logData)
	}
}

func TestGetLogGoodMockByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodMockByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodMockByChar.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodMockByChar")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("mock", 0, math.MaxInt64, true)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "Mock") {
		t.Error("Expecting 'Mock' in reponse but gets:" + logData)
	}
}

func TestGetLogUnexistantState(t *testing.T) {
	t.Log("Entering... TestGetLogUnexistantState")
	statesPath := "../../test/resource/states-TestGetLogUnexistantState.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogUnexistantState")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("UnexistantState", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Should throw an error as the state doesn't exist")
	}
	t.Log(err.Error())
}

func TestGetLogEmptyState(t *testing.T) {
	t.Log("Entering... TestGetLogUnexistantState")
	statesPath := "../../test/resource/states-TestGetLogEmptyState.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogEmptyStat")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Should throw an error as the state is empty")
	}
	t.Log(err.Error())
}

func TestGetLogGoodFromToInRange(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToInRange")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToInRange.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodFromToInRange")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 2, 2, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	t.Log("Log:\n" + logData)
	if strings.Contains(logData, "sample-state1 line 1") ||
		strings.Contains(logData, "sample-state1 line 4") ||
		strings.Contains(logData, "sample-state1 line 5") ||
		!strings.Contains(logData, "sample-state1 line 2") ||
		!strings.Contains(logData, "sample-state1 line 3") {
		t.Error("Expecting 'state1' in reponse but gets:" + logData)
	}
}

func TestGetLogGoodFromToInRangeByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToInRangeByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToInRangeByChar.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodFromToInRangeByChar")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 2, 2, true)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if logData != "mp" {
		t.Error("Expecting 'mp' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodFromToOutOfRange(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToOutOfRange")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToOutOfRange.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodFromToOutOfRange")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 6, 7, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if logData != "" {
		t.Error("Expecting empty in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodFromToOutOfRangeByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToOutOfRangeByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToOutOfRangeByChar.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogGoodFromToOutOfRangeByChar")
	sm.StatesPath = statesPath
	raw, err := sm.GetLog("state1", 6000, 7, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if logData != "" {
		t.Error("Expecting empty in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogStateNotExists(t *testing.T) {
	t.Log("Entering... TestGetLogStateNotExists")
	statesPath := "../../test/resource/states-TestGetLogStateNotExists.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogStateNotExists")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("state1", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetLogStateMalformed(t *testing.T) {
	t.Log("Entering... TestGetLogStateMalformed")
	statesPath := "../../test/resource/states-malformed.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-malformed")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("state1", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("expecting error as the states file is malformed")
	}
	t.Log(err.Error())
}

func TestGetLogNoLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogNoLogPath")
	statesPath := "../../test/resource/states-TestGetLogNoLogPath.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogNoLogPath")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("nologpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as states files doesn't provide log path.")
	}
	t.Log(err.Error())
}

func TestGetLogEmptyLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogEmptyLogPath")
	statesPath := "../../test/resource/states-TestGetLogEmptyLogPath.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogEmptyLogPath")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("emptylogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as the log path is empty")
	}
	t.Log(err.Error())
}

func TestGetLogNilLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogNilLogPath")
	statesPath := "../../test/resource/states-TestGetLogNilLogPath.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogNilLogPath")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("nillogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as the log path is nil")
	}
	t.Log(err.Error())
}

func TestGetLogWrongLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogWrongLogPath")
	statesPath := "../../test/resource/states-TestGetLogWrongLogPath.yaml"
	// sm, err := newStateManager(statesPath)
	sm := newStateManager("states-TestGetLogWrongLogPath")
	sm.StatesPath = statesPath
	_, err := sm.GetLog("wronglogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as it has a wring log path")
	}
	t.Log(err.Error())
}
