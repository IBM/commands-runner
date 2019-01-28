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
package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	yaml "gopkg.in/yaml.v2"
)

//This program runs a state file provided in the first parameter.
func ResetStateManager(extensionName string) {

	log.SetLevel(log.DebugLevel)
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
	states, err := stateManagerInstance.GetStates("", false, true)
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
	states, err = stateManagerInstance.GetStates("", false, true)
	statesOut, err = yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

}
