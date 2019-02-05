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
