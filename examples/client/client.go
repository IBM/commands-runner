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
	"os"

	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunnerCLI"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/client/helloWorld"
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
