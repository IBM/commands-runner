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
	"os"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/IBM/commands-runner/api/commandsRunnerCLI"
	"github.com/IBM/commands-runner/examples/client/helloWorld"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func main() {
	//Define a dummy function
	helloWorld := func(c *cli.Context) error {
		client, errClient := helloWorld.NewClient(commandsRunnerCLI.URL, commandsRunnerCLI.OutputFormat, commandsRunnerCLI.Timeout, commandsRunnerCLI.CACertPath, commandsRunnerCLI.InsecureSSL, commandsRunnerCLI.Token, commandsRunnerCLI.DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.HelloWorld()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	//Create a commandsRunner client.
	app := commandsRunnerCLI.Client()

	//Overwrite some app parameters
	app.Usage = "client ..."
	app.Version = "1.0.0"
	app.Description = "Sample client"

	//Enrich with extra client commands
	app.Commands = append(app.Commands, []cli.Command{
		{
			Name:   "hello",
			Usage:  "hello",
			Action: helloWorld,
		},
	}...)

	//Run the command
	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
