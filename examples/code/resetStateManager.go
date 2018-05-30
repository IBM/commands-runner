package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/stateManager"
)

//This program runs a state file provided in the first parameter.
func (t T) ResetStateManager(statesPath string) {

	//Create a new stateManagerInstance
	stateManagerInstance, err := stateManager.NewStateManager(statesPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	//Print states
	states, err := stateManagerInstance.GetStates("")
	statesOut, err := yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

	//Reset all states to READY to run it again
	err = stateManagerInstance.ResetEngine()
	if err != nil {
		log.Error(err.Error())
	}

	//Print states
	states, err = stateManagerInstance.GetStates("")
	statesOut, err = yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

}
