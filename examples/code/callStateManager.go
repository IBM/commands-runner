package main

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/stateManager"
)

//This program runs a state file provided in the first parameter.
func CallStateManager(statesPath string) {

	//Create a new stateManagerInstance
	stateManagerInstance, err := state.NewStateManager(statesPath)
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

	//Run the state file
	err = stateManagerInstance.Start()
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

	//Get logs
	logs, err := stateManagerInstance.GetLogs(0, math.MaxInt64, true)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(logs)

}
