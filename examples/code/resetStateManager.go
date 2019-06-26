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
package main

import (
	"fmt"

	"github.com/IBM/commands-runner/api/commandsRunner/state"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

//This program runs a state file provided in the first parameter.
func ResetStateManager(extensionName string) {

	// log.SetLevel(log.DebugLevel)
	log.Info("============== Init Extensions ========================")
	//This initializes the stateManager and register all embedded extensions mentioned in the examples/extensions/test-extensions.yml

	state.InitExtensions("examples/extensions/test-extensions.yml", "examples/extensions", "examples/data/extensions/", "")

	//Create a new stateManagerInstance
	log.Info("============== Create new StatesManager ========================")
	stateManagerInstance, err := state.GetStatesManager(extensionName)
	if err != nil {
		log.Fatal(err.Error())
	}

	//Print states
	log.Info("============== Print states ========================")
	states, err := stateManagerInstance.GetStates("", false, true, nil)
	statesOut, err := yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

	//Reset all states to READY to run it again
	log.Info("============== Reset status ========================")
	err = stateManagerInstance.ResetEngine()
	if err != nil {
		log.Error(err.Error())
	}

	//Print states
	log.Info("============== Print states ========================")
	states, err = stateManagerInstance.GetStates("", false, true, nil)
	statesOut, err = yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

}
