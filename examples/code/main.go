package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

var statesFilePath string

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
					Name:        "statesFilePath, s",
					Usage:       "States file path",
					Destination: &statesFilePath,
				},
			},
			Action: func(c *cli.Context) error {
				CallStateManager(statesFilePath)
				return nil
			},
		},
		{
			Name:  "ResetStateManager",
			Usage: "Reset a state file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "statesFilePath, s",
					Usage:       "States file path",
					Destination: &statesFilePath,
				},
			},
			Action: func(c *cli.Context) error {
				ResetStateManager(statesFilePath)
				return nil
			},
		},
	}

	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
