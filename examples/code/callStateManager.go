package main

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	yaml "gopkg.in/yaml.v2"
)

//This program runs a state file provided in the first parameter.
func CallStateManager(extensionName string) {
	log.SetLevel(log.DebugLevel)
	log.Info("============== Init Extensions ========================")
	//This initializes the stateManager and register all embedded extensions mentioned in the examples/extensions/test-extensions.yml

	state.InitExtensions("examples/extensions/test-extensions.yml", "examples/extensions", "examples/data/extensions/", "../../../logs")

	log.Info("============== Create new StatesManager ========================")
	//Create a new stateManagerInstance
	stateManagerInstance, err := state.GetStatesManager(extensionName)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("============== Print states ========================")
	//Print states
	states, err := stateManagerInstance.GetStates("", false, true)
	statesOut, err := yaml.Marshal(states)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(string(statesOut))

	log.Info("============== Run ========================")
	//Run the state file
	err = stateManagerInstance.Start()
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

	log.Info("============== Show logs ========================")
	//Get logs
	logs, err := stateManagerInstance.GetLogs(0, math.MaxInt64, true)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(logs)

}
