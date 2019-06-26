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
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

var extensionName string

func main() {
	app := cli.NewApp()
	app.Usage = "Commands Runner for installation"
	app.Description = "CLI to manage initial Commands Runner installation"

	app.Commands = []cli.Command{
		{
			Name:  "CallStateManager",
			Usage: "Call the state manager on a given state file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extensionName, e",
					Usage:       "Extension Name",
					Destination: &extensionName,
				},
			},
			Action: func(c *cli.Context) error {
				CallStateManager(extensionName)
				return nil
			},
		},
		{
			Name:  "ResetStateManager",
			Usage: "Reset a state file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extensionName, e",
					Usage:       "Extension Name",
					Destination: &extensionName,
				},
			},
			Action: func(c *cli.Context) error {
				ResetStateManager(extensionName)
				return nil
			},
		},
	}

	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
