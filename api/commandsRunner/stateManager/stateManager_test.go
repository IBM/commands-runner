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
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"

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
	level, _ := log.ParseLevel("debug")
	log.SetLevel(level)
	directorlog := "sample-director line 1\nsample-director line 2\nsample-director line 3\nsample-director line 4\nsample-director line 5"
	cflog := "sample-cf line 1\nsample-cf line 2\nsample-cf line 3\nsample-cf line 4\nsample-cf line 5"
	ioutil.WriteFile("/tmp/sample-director.log", []byte(directorlog), 0600)
	errDirector := os.Chmod("/tmp/sample-director.log", 0600)
	if errDirector != nil {
		log.Fatal(errDirector.Error())
	}
	ioutil.WriteFile("/tmp/sample-cf.log", []byte(cflog), 0600)
	errCF := os.Chmod("/tmp/sample-cf.log", 0600)
	if errCF != nil {
		log.Fatal(errCF.Error())
	}
}

func TestGetStatesOK(t *testing.T) {
	t.Log("Entering... TestGetStatesOK")
	statesPath := "../../test/resource/states-TestGetStatesOK.yaml"
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	states, err := sm.GetStates("")
	if err != nil {
		t.Error(err.Error())
	}
	stateM, _ := json.Marshal(states)
	var statesIn States
	statesData, err := ioutil.ReadFile(statesPath)
	t.Logf("StatesData=%v", statesData)
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
	t.Log(string(expected))
	t.Log(string(got))
	if string(got) != string(expected) {
		t.Errorf("handler returned unexpected response: got %v want %v",
			string(stateM), string(expected))
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetStatesWithStatus(t *testing.T) {
	t.Log("Entering... TestGetStatesWithStatus")
	sm, err := NewStateManager("../../test/resource/states-run-running.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	states, err := sm.GetStates(StateRUNNING)
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-post-sample-from-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-states-post-sample-from-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON := "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	t.Log("Entering... TestSetStatesOK")
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStatesStatusesOK(t *testing.T) {
	t.Log("Entering... TestSetStatesStatusesOK")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	sm, err := NewStateManager("../../test/resource/states-TestSetStatesStatusesOK.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.SetStatesStatuses("SKIP", "director", true, "director", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err := sm.GetState("director")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status != StateSKIP {
		t.Error("Status not set as SKIP as expected:" + state.Status)
	}
	states, err := sm.GetStates(StateSKIP)
	if len(states.StateArray) > 1 {
		t.Error("Another state was set to SKIP")
	}
	err = sm.SetStatesStatuses("READY", "director", true, "director", true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStatesStatusesFromTo(t *testing.T) {
	t.Log("Entering... TestSetStatesStatusesOK")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	sm, err := NewStateManager("../../test/resource/states-TestSetStatesStatusesFromTo.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	//Test a range in the middle inclusive
	err = sm.SetStatesStatuses("SKIP", "repeat", true, "nologpath", true)
	if err != nil {
		t.Error(err.Error())
	}
	state, err := sm.GetState("director")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP as expected:" + state.Status)
	}
	states, err := sm.GetStates(StateSKIP)
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
	state, err = sm.GetState("director")
	if err != nil {
		t.Error(err.Error())
	}
	if state.Status == StateSKIP {
		t.Error("Status set as SKIP and got :" + state.Status)
	}
	states, err = sm.GetStates(StateSKIP)
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
	states, err = sm.GetStates(StateSKIP)
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
	states, err = sm.GetStates(StateSKIP)
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
	states, err = sm.GetStates(StateSKIP)
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
	states, err = sm.GetStates(StateSKIP)
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
	err = sm.SetStatesStatuses("SKIP", "repeat", true, "director", true)
	if err == nil {
		t.Error("Range in wrong order error expected")
	}
	t.Log(err.Error())
	err = sm.SetStatesStatuses("READY", "", true, "", true)
	if err != nil {
		t.Error(err.Error())
	}
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
	sm, err := NewStateManager("../../test/resource/out-test-states-delete-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON := "{\"states\":[{\"name\":\"director\",\"delete:\":true,\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	t.Log("Entering... TestSetStatesOK")
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("director")
	if err != nil {
		t.Error(err.Error())
	}
	states = *statesResult
	if len(states.StateArray) > 0 {
		t.Error("Director not removed")
	}
	statesResult, err = sm.GetStates("cf")
	if err != nil {
		t.Error(err.Error())
	}
	states = *statesResult
	if len(states.StateArray) > 0 {
		t.Error("CF removed")
	}
}

func TestSetStatesMerge(t *testing.T) {
	t.Log("Entering... TestSetStatesMerge")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-merge-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-states-merge-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON := "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	statesData, _ := sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf2\",\"label\":\"CloudFoundry2\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, false)
	statesData, _ = sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("")
	if err != nil {
		t.Error(err.Error())
	}
	if statesResult.StateArray[0].Name != "director" &&
		!(statesResult.StateArray[1].Name == "cf" || statesResult.StateArray[1].Name == "cf2") &&
		!(statesResult.StateArray[1].Name == "cf2" || statesResult.StateArray[1].Name == "cf") {
		t.Error("Wrong order")
	}
}

func TestSetStatesMergeWithDelete(t *testing.T) {
	t.Log("Entering... TestSetStatesMergeWithDelete")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-merge-delete-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-states-merge-delete-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON := "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf2\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"deleted\":true,\"label\":\"CloudFoundry2\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, false)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	statesResult, err := sm.GetStates("")
	if err != nil {
		t.Error(err.Error())
	}
	if len(statesResult.StateArray) != 2 {
		t.Error("Result with size " + strconv.Itoa(len(statesResult.StateArray)) + " and expect 2")
	}
	if statesResult.StateArray[0].Name != "director" &&
		statesResult.StateArray[1].Name != "cf2" {
		t.Error("Wrong order")
	}
}

func TestSetStatesMergeWithCycle(t *testing.T) {
	t.Log("Entering... TestSetStatesMergeWithCycle")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-states-merge-cycle-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-states-merge-cycle-json.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON := "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	var states States
	json.Unmarshal([]byte(statesJSON), &states)
	err = sm.SetStates(states, true)
	if err != nil {
		t.Error(err.Error())
	}
	statesJSON = "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\",\"next_states\":[\"director\"]}]}"
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
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestInsertStatesBeforeFirst(t *testing.T) {
	t.Log("Entering... TestInsertStatesBeforeFirst")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-before-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-before-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
		Name: "cfp-ext-template",
	}
	err = sm.InsertState(*state, 1, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "cfp-ext-template" ||
		sm.StateArray[1].Name != "First" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesAfterFirst(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-after-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	t.Log("cfp-ext-template")
	state := &State{
		Name: "cfp-ext-template",
	}
	err = sm.InsertState(*state, 1, "", false)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "cfp-ext-template" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
}
func TestInsertStatesBeforeLast(t *testing.T) {
	t.Log("Entering... TestInsertStatesBeforeLast")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-before-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-before-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
		Name: "cfp-ext-template",
	}
	err = sm.InsertState(*state, 2, "", true)
	stateData, _ = sm.convert2String()
	t.Log(stateData)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "cfp-ext-template" ||
		sm.StateArray[2].Name != "Last" {
		t.Error("Not inserted at the correct position")
	}
	//	t.Error("")
}

func TestInsertStatesAfterLast(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	t.Log("cfp-ext-template")
	state := &State{
		Name: "cfp-ext-template",
	}
	err = sm.InsertState(*state, 2, "", false)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "Last" ||
		sm.StateArray[2].Name != "cfp-ext-template" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesAfterLastByName(t *testing.T) {
	t.Log("Entering... TestInsertStatesAfterFirst")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-after-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	t.Log("cfp-ext-template")
	state := &State{
		Name: "cfp-ext-template",
	}
	err = sm.InsertState(*state, 0, "Last", false)
	statesData, _ = sm.convert2String()
	t.Log(statesData)
	if err != nil {
		t.Error(err.Error())
	}
	if sm.StateArray[0].Name != "First" ||
		sm.StateArray[1].Name != "Last" ||
		sm.StateArray[2].Name != "cfp-ext-template" {
		t.Error("Not inserted at the correct position")
	}
}

func TestInsertStatesWithCycle(t *testing.T) {
	t.Log("Entering... TestInsertStatesWithCycle")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-insert-cycle-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-insert-cycle-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
		Name: "cfp-ext-template",
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	t.Log("Entering... TestInsertStatesAfterFirst")
	_, err := os.Create("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-delete-first-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	t.Log("Entering... TestInsertStatesAfterFirst")
	_, err := os.Create("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	sm, err := NewStateManager("../../test/resource/out-test-delete-last-state.yaml")
	if err != nil {
		t.Error(err.Error())
	}
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
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	sm, err := NewStateManager("../../test/resource/states-TestGetStateOK.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	state, err := sm.GetState("director")
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
	sm, err := NewStateManager("../../test/resource/states-TestGetStateNOK.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	stateN, err := sm.GetState("not-exist")
	if err == nil {
		t.Error("Found a state but should'nt " + stateN.Status)
	}
}

func TestGetStateEmptyState(t *testing.T) {
	t.Log("Entering... TestGetStateNOK")
	sm, err := NewStateManager("../../test/resource/states-TestGetStateEmptyState.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetState("")
	if err == nil {
		t.Error("An error should be raised as the state is not specified")
	}
}

func TestSetStateStatus(t *testing.T) {
	t.Log("Entering... TestSetStateStatus")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	stateV := StateRUNNING
	reasonV := strconv.FormatInt(time.Now().UnixNano(), 5)
	scriptV := strconv.FormatInt(time.Now().UnixNano(), 7)
	scriptTimeoutV := time.Now().Second()
	statesPath := "../../test/resource/states-TestSetStateStatus.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.SetState("cf", stateV, reasonV, scriptV, scriptTimeoutV, true)
	if err != nil {
		t.Error(err.Error())
	}
	stateN, err := sm.GetState("cf")
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
	err = sm.SetState("cf", StateFAILED, reasonV, scriptV, scriptTimeoutV, true)
	if err != nil {
		t.Error(err.Error())
	}
	stateN, err = sm.GetState("cf")
	if err != nil {
		t.Error(err.Error())
	}
	if stateN.Reason != reasonV {
		t.Error("Expected:" + reasonV + " gets:" + stateN.Reason)
	}
	err = sm.SetState("cf", StateFAILED, "", "hello.sh", 61, true)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.SetState("cf", "WrongStatus", "", "hello.sh", 61, true)
	if err == nil {
		t.Error("Expecting error as the status is incorrect")
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineSuccess(t *testing.T) {
	t.Log("Entering... TestEngineSuccess")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-success.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Reset States file")
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates(StateFAILED)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) > 0 {
		t.Error("At least one state failed:" + states.StateArray[0].Name)
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineWithRerunAfter(t *testing.T) {
	t.Log("Entering... TestEngineWithRerunAfter")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-with-rerun-after.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Reset States file")
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates(StateFAILED)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) > 0 {
		t.Error("At least one state failed:" + states.StateArray[0].Name)
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineWithRerunBefore(t *testing.T) {
	t.Log("Entering... TestEngineWithRerunAfter")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-with-rerun-before.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Reset States file")
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	err = sm.Execute("task1", "task3")
	if err == nil {
		t.Error("Expect error because the stateToRerun of task2 reference task1 which is before in the sequence")
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineCycleOnRerun(t *testing.T) {
	t.Log("Entering... TestEngineSuccess")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-cycle-on-rerun.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	err = sm.Start()
	if err == nil {
		t.Error("Expected error as it has a cyle task1->task2")
	} else {
		t.Log(err.Error())
	}
}

func TestEngineCycleOnNext(t *testing.T) {
	t.Log("Entering... TestEngineSuccess")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-cycle-on-next.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	err = sm.Start()
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
	sm.setDefaultValues()
	statesData, _ := sm.convert2String()
	t.Log(statesData)
	if len(sm.StateArray[0].NextStates) == 0 {
		t.Error("next_states not updated")
	}
	t.Logf("%+v", sm.StateArray[0].NextStates)
}
func TestStatusFailedDependency(t *testing.T) {
	t.Log("Entering... TestStatusFailedDependency")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-failed-dependency.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Reset States file")
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Execute states file")
	sm.Execute("task1", "task3")
	t.Log("Get Failed states")
	states, errStates := sm.GetStates("")
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if states.StateArray[0].Status != StateSUCCEEDED ||
		states.StateArray[1].Status != StateFAILED ||
		states.StateArray[2].Status != StateFAILED {
		t.Error("The statuses are not correct, expecting task1: SUCCEEDED, task2: FAILED, task3: FAILED")
	}
	t.Log(states)
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEngineFailure(t *testing.T) {
	t.Log("Entering... TestEngineFailure")
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-run-failure.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Reset States file")
	err = sm.ResetEngine()
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
	states, errStates := sm.GetStates(StateFAILED)
	if errStates != nil {
		t.Error(errStates.Error())
	}
	if len(states.StateArray) <= 0 {
		t.Error("At least one state must failed")
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSetStateStatusEmptyState(t *testing.T) {
	extensionManager.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../../test/data/extensions/")
	statesPath := "../../test/resource/states-TestSetStateStatusEmptyState.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("Entering... TestSetStateStatus")
	err = sm.SetState("", "TEST", "REASON", "", -1, true)
	if err == nil {
		t.Error("An error should be raised as the state is not specified")
	}
	err = sm.ResetEngine()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetLogGoodDirector(t *testing.T) {
	t.Log("Entering... TestGetLogGoodDirector")
	statesPath := "../../test/resource/states-TestGetLogGoodDirector.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "director") {
		t.Error("Expecting 'director' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodDirectorByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodDirectorByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodDirectorByChar.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 0, math.MaxInt64, true)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "director") {
		t.Error("Expecting 'director' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodCF(t *testing.T) {
	t.Log("Entering... TestGetLogGoodCF")
	statesPath := "../../test/resource/states-TestGetLogGoodCF.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("cf", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	if !strings.Contains(logData, "cf") {
		t.Error("Expecting 'cf' in reponse but gets:" + logData)
	}
	t.Log(string(raw))
}

func TestGetLogGoodMock(t *testing.T) {
	t.Log("Entering... TestGetLogGoodMock")
	statesPath := "../../test/resource/states-TestGetLogGoodMock.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
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
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
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
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("UnexistantState", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Should throw an error as the state doesn't exist")
	}
	t.Log(err.Error())
}

func TestGetLogEmptyState(t *testing.T) {
	t.Log("Entering... TestGetLogUnexistantState")
	statesPath := "../../test/resource/states-TestGetLogEmptyState.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Should throw an error as the state is empty")
	}
	t.Log(err.Error())
}

func TestGetLogGoodFromToInRange(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToInRange")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToInRange.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 2, 2, false)
	if err != nil {
		t.Error(err.Error())
	}
	logData := string(raw)
	t.Log("Log:\n" + logData)
	if strings.Contains(logData, "sample-director line 1") ||
		strings.Contains(logData, "sample-director line 4") ||
		strings.Contains(logData, "sample-director line 5") ||
		!strings.Contains(logData, "sample-director line 2") ||
		!strings.Contains(logData, "sample-director line 3") {
		t.Error("Expecting 'director' in reponse but gets:" + logData)
	}
}

func TestGetLogGoodFromToInRangeByChar(t *testing.T) {
	t.Log("Entering... TestGetLogGoodFromToInRangeByChar")
	statesPath := "../../test/resource/states-TestGetLogGoodFromToInRangeByChar.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 2, 2, true)
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
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 6, 7, false)
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
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	raw, err := sm.GetLog("director", 6000, 7, false)
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
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("director", 0, math.MaxInt64, false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetLogStateMalformed(t *testing.T) {
	t.Log("Entering... TestGetLogStateMalformed")
	statesPath := "../../test/resource/states-malformed.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("director", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("expecting error as the states file is malformed")
	}
	t.Log(err.Error())
}

func TestGetLogNoLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogNoLogPath")
	statesPath := "../../test/resource/states-TestGetLogNoLogPath.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("nologpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as states files doesn't provide log path.")
	}
	t.Log(err.Error())
}

func TestGetLogEmptyLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogEmptyLogPath")
	statesPath := "../../test/resource/states-TestGetLogEmptyLogPath.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("emptylogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as the log path is empty")
	}
	t.Log(err.Error())
}

func TestGetLogNilLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogNilLogPath")
	statesPath := "../../test/resource/states-TestGetLogNilLogPath.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("nillogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as the log path is nil")
	}
	t.Log(err.Error())
}

func TestGetLogWrongLogPath(t *testing.T) {
	t.Log("Entering... TestGetLogWrongLogPath")
	statesPath := "../../test/resource/states-TestGetLogWrongLogPath.yaml"
	sm, err := NewStateManager(statesPath)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = sm.GetLog("wronglogpath", 0, math.MaxInt64, false)
	if err == nil {
		t.Error("Expecting error as it has a wring log path")
	}
	t.Log(err.Error())
}
