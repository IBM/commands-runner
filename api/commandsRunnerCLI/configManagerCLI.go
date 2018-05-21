/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package commandsRunnerCLI

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/resourceManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunnerCLI/clientConfigManager"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func main() {
	var url string
	var outputFormat string
	var timeout string
	var caCertPath string
	var insecureSSL bool
	var token string
	var state string
	var newStatus string
	var stateTimeout string
	var searchStatus string
	var fromState, toState string
	var extensionName string
	var tokenOutputFilePath string
	var extensionsToList string
	var extensionZipPath string
	var statesPath string
	var statePath string
	var insertExtensionName string
	var statePosition int
	var stateName string

	var curlMethod string
	var curlDataPath string

	var logLevel string

	var statusName string
	var status string

	getStatus := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetCMStatus()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	setStatus := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetCMStatus(statusName, status)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	getAPI := func(c *cli.Context) error {
		data, errClient := configManagerClient.GetAPISetup(outputFormat)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		fmt.Println(data)
		return nil
	}

	setAPI := func(c *cli.Context) error {
		errClient := configManagerClient.SetAPISetup(url, outputFormat, timeout, caCertPath, token)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		return nil
	}

	removeAPI := func(c *cli.Context) error {
		errClient := configManagerClient.RemoveAPISetup()
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		return nil
	}

	createToken := func(c *cli.Context) error {
		if tokenOutputFilePath != "" {
			return configManagerClient.NewTokenFile(tokenOutputFilePath)
		}
		data := configManagerClient.NewToken()
		fmt.Print(data)
		return nil
	}

	getBOM := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetBOM()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	getLogs := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		err := client.GetLogs(extensionName, state, c.Bool("follow"), c.Bool("quiet"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	}

	getUIConfig := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, errClient := client.GetUIConfig(extensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		fmt.Println(data)
		return nil
	}

	deploy := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.StartEngine(extensionName, fromState, toState)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		if c.Bool("wait") {
			err := client.GetLogs(extensionName, state, true, false)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		}
		return nil
	}

	isRunning := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.IsRunningEngine(extensionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	reset := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		_, err := client.ResetEngine(extensionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	}

	setState := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		err := client.SetState(extensionName, state, newStatus, stateTimeout)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	}

	getState := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetState(extensionName, state)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	getStates := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetStates(extensionName, "")
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	setStates := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetStates(extensionName, statesPath, true)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	setStatesStatuses := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetStatesStatuses(extensionName, newStatus, fromState, c.Bool("from-included"), toState, c.Bool("to-included"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	mergeStates := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetStates(extensionName, statesPath, false)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	insertStateStates := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		// if statePath == "" {
		// 	err := errors.New("statePath must be provided")
		// 	fmt.Println(err.Error())
		// 	return err
		// }
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.InsertStateStates(extensionName, statePosition, stateName, c.Bool("before"), statePath, insertExtensionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	deleteStateStates := func(c *cli.Context) error {
		if extensionName == "" {
			extensionName = global.CloudFoundryPieName
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.DeleteStateStates(extensionName, statePosition, stateName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	findStates := func(c *cli.Context) error {
		if searchStatus == "" {
			err := errors.New("--status, -s missing")
			fmt.Println(err.Error())
			return err
		}
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetStates(extensionName, searchStatus)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	register := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.RegisterExtension(extensionZipPath, extensionName, c.Bool("force"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Printf(data)
		return nil
	}

	unregister := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.UnregisterExtension(extensionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Printf(data)
		return nil
	}

	getExtensions := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetExtensions(extensionsToList, c.Bool("catalog"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	curl := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		if len(c.Args()) != 1 {
			return errors.New("Invalid number of parameters")
		}
		data, err := client.Curl(curlMethod, c.Args()[0], curlDataPath)
		fmt.Print(data)
		return err
	}

	getPCMLogLevel := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetPCMLogLevel()
		fmt.Print(data)
		return err
	}

	setPCMLogLevel := func(c *cli.Context) error {
		client, errClient := configManagerClient.NewClient(url, outputFormat, timeout, insecureSSL)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetPCMLogLevel(logLevel)
		fmt.Print(data)
		return err
	}

	app := cli.NewApp()
	app.Usage = "Config Manager for Cloud Foundry installation"
	raw, e := resourceManager.Asset("VERSION")
	if e != nil {
		log.Fatal("Version not found")
	}
	app.Version = string(raw)
	app.Description = "CLI to manage initial Cloud Foundry installation"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, u",
			Usage:       "API Url",
			Destination: &url,
		},
		cli.StringFlag{
			Name:        "format, f",
			Usage:       "Output format (text, json, yaml)",
			Destination: &outputFormat,
		},
		cli.StringFlag{
			Name:        "timeout, t",
			Usage:       "RestAPI timeout in seconds",
			Destination: &timeout,
		},
		cli.StringFlag{
			Name:        "cacert",
			Usage:       "CA Cert path",
			Destination: &caCertPath,
		},
		cli.BoolFlag{
			Name:        "insecure, k",
			Usage:       "Turn off verification",
			Destination: &insecureSSL,
		},
		cli.StringFlag{
			Name:        "token",
			Usage:       "Token",
			Destination: &token,
		},
	}

	app.Commands = []cli.Command{
		/*            API                  */
		{
			Name:   "api",
			Usage:  "API endpoint management",
			Action: getAPI,
			Subcommands: []cli.Command{
				{
					Name:    "save",
					Aliases: []string{"s"},
					Usage:   "Save API setup",
					Action:  setAPI,
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "Remove API setup",
					Action:  removeAPI,
				},
			},
		},
		/*            TOKEN                  */
		{
			Name:  "token",
			Usage: "Token management",
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create token",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "outFilePath, o",
							Usage:       "Output file",
							Destination: &tokenOutputFilePath,
						},
					},
					Action: createToken,
				},
			},
		},
		/*            Extensions                  */
		{
			Name:  "extensions",
			Usage: "List extensions",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "list, l",
					Usage:       "Which extensions to list (IBM or custom), leave empty for all",
					Destination: &extensionsToList,
				},
				cli.BoolFlag{
					Name:  "catalog, c",
					Usage: "Display unregistered IBM extensions",
				},
			},
			Action: getExtensions,
		},
		/*            EXTENSION                  */
		{
			Name:  "extension",
			Usage: "Extension management",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "uiconfig",
					Usage:  "Get the uiconfig for extension",
					Action: getUIConfig,
				},
				{
					Name:  "register",
					Usage: "Register an extension",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "path, p",
							Usage:       "Path to extension being registered",
							Destination: &extensionZipPath,
						},
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force registration however the extension is already registered",
						},
					},
					Action: register,
				},
				{
					Name:    "deploy",
					Aliases: []string{"d"},
					Usage:   "Deploy extension",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "fromState, f",
							Usage:       "Start from the provided state",
							Destination: &fromState,
						},
						cli.StringFlag{
							Name:        "toState, t",
							Usage:       "Finish at the provided state included",
							Destination: &toState,
						},
						cli.BoolFlag{
							Name:  "wait, w",
							Usage: "Wait until deployment ends",
						},
					},
					Action: deploy,
				},
				{
					Name:    "unregister",
					Aliases: []string{"u"},
					Usage:   "Remove an existing extension",
					Action:  unregister,
				},
				{
					Name:    "reset",
					Aliases: []string{"r"},
					Usage:   "Reset the engine, all statuses not equal SKIP will be set to READY",
					Action:  reset,
				},
				{
					Name:        "logs",
					Aliases:     []string{"l"},
					Usage:       "Display logs for a given state, if no state provided then the current running state log is returned",
					Description: "Command to retrieve deployment logs",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "state, s",
							Usage:       "State",
							Destination: &state,
						},
						cli.BoolFlag{
							Name:  "follow, f",
							Usage: "Continuously follow the log",
						},
						cli.BoolFlag{
							Name:  "quiet, q",
							Usage: "Don't display logs but return control when deployment finished",
						},
					},
					Action: getLogs,
				},
			},
		},
		/*            UICONFIG                  */
		{
			Name:  "uiconfig",
			Usage: "UIConfig management",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "config, c",
					Usage:       "Configuration name",
					Destination: &extensionName,
				},
			},
			Action: getUIConfig,
		},
		/*            CFP                  */
		{
			Name:  "cfp",
			Usage: "cfp information",
			Subcommands: []cli.Command{
				{
					Name:    "bom",
					Aliases: []string{"b"},
					Usage:   "Get the bom",
					Action:  getBOM,
				},
			},
		},
		/*            PCM                  */
		{
			Name:  "pcm",
			Usage: "pcm management",
			Subcommands: []cli.Command{
				{
					Name:   "log-level",
					Usage:  "Get the current log level of the platform config manager",
					Action: getPCMLogLevel,
				},
				{
					Name:   "set-log-level",
					Usage:  "Get the current log level of the platform config manager",
					Action: setPCMLogLevel,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "level, l",
							Usage:       "Requested log level",
							Destination: &logLevel,
						},
					},
				},
			},
		},
		/*            curl                  */
		{
			Name:   "curl",
			Usage:  "curl a url",
			Action: curl,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "method, X",
					Usage:       "Curl method",
					Destination: &curlMethod,
				},
				cli.StringFlag{
					Name:        "data, d",
					Usage:       "The file to send as data",
					Destination: &curlDataPath,
				},
			},
		},
		/*            Deployment                  */
		{
			Name:  "engine",
			Usage: "Manage engine (start, reset, isRunning)",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
			},
			Action: isRunning,
			Subcommands: []cli.Command{
				{
					Name:    "start",
					Aliases: []string{"s"},
					Hidden:  true,
					Usage:   "Start the engine",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "fromState, f",
							Usage:       "Start from the provided state",
							Destination: &fromState,
						},
						cli.StringFlag{
							Name:        "toState, t",
							Usage:       "Finish at the provided state included",
							Destination: &toState,
						},
					},
					Action: deploy,
				},
				{
					Name:    "reset",
					Aliases: []string{"r"},
					Usage:   "Reset the engine, all statuses not equal SKIP will be set to READY",
					Action:  reset,
				},
			},
		},
		/*            LOGS                  */
		{
			Name:        "logs",
			Aliases:     []string{"l"},
			Usage:       "Display logs for a given state, if no state provided then the current running state log is returned",
			Description: "Command to retrieve deployment logs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
				cli.StringFlag{
					Name:        "state, s",
					Usage:       "State",
					Destination: &state,
				},
				cli.BoolFlag{
					Name:  "follow, f",
					Usage: "Continuously follow the log",
				},
				cli.BoolFlag{
					Name:   "quiet, q",
					Hidden: true,
					Usage:  "Don't display logs but return control when deployment finished",
				},
			},
			Action: getLogs,
		},
		/*            CM STATUS                  */
		{
			Name:   "status",
			Usage:  "Get/Set Config Manager statuses",
			Action: getStatus,
			Subcommands: []cli.Command{
				{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "Set status",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "name, n",
							Usage:       "Status name",
							Destination: &statusName,
						},
						cli.StringFlag{
							Name:        "status, s",
							Usage:       "Status value",
							Destination: &status,
						},
					},
					Action: setStatus,
				},
			},
		},
		/*            STATE                  */
		{
			Name:      "state",
			Usage:     "Manage a given state (get, set)",
			UsageText: "cm state [-e <extension_name>] -s <state> set --status <new_status>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
				cli.StringFlag{
					Name:        "state, s",
					Usage:       "State",
					Destination: &state,
				},
			},
			Action: getState,
			Subcommands: []cli.Command{
				{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "Set status/timeout of a given state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "status",
							Usage:       "Status to set",
							Destination: &newStatus,
						},
						cli.StringFlag{
							Name:        "timeout",
							Usage:       "Timeout to set",
							Destination: &stateTimeout,
						},
					},
					Action: setState,
				},
			},
		},
		/*            STATES                  */
		{
			Name:  "states",
			Usage: "Manage states (get, find)",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
			},
			Action: getStates,
			Subcommands: []cli.Command{
				{
					Name:    "find",
					Aliases: []string{"f"},
					Usage:   "Find the states with a given status",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "status, s",
							Usage:       "Status to find",
							Destination: &searchStatus,
						},
					},
					Action: findStates,
				},
				{
					Name:    "insert",
					Aliases: []string{"i"},
					Usage:   "Insert a state into the state file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "state-path, s",
							Usage:       "State Path",
							Destination: &statePath,
						},
						cli.StringFlag{
							Name:        "insert-extension-name, i",
							Usage:       "Name of the extension to inssert",
							Destination: &insertExtensionName,
						},
						cli.StringFlag{
							Name:        "state-name, n",
							Usage:       "Referencing state name, can be use instead of position",
							Destination: &stateName,
						},
						cli.IntFlag{
							Name:        "position, p",
							Usage:       "Position where to insert the state, you can also use -n to specify the state name to use are reference",
							Destination: &statePosition,
						},
						cli.BoolFlag{
							Name:  "before, b",
							Usage: "Insert the state before the provided position/state name",
						},
					},
					Action: insertStateStates,
				},
				{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "Delete a state from the state file",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:        "position, p",
							Usage:       "Position where to insert the state",
							Destination: &statePosition,
						},
						cli.StringFlag{
							Name:        "state-name, n",
							Usage:       "Referencing state name, can be use instead of position",
							Destination: &stateName,
						},
					},
					Action: deleteStateStates,
				},
				{
					Name:    "set",
					Aliases: []string{"s"},
					Hidden:  true,
					Usage:   "set a states file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "states, s",
							Usage:       "States file path to set",
							Destination: &statesPath,
						},
					},
					Action: setStates,
				},
				{
					Name:    "set-status-by-range",
					Aliases: []string{"s"},
					Usage:   "set states statuses by providing a from/to state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "status",
							Usage:       "New status to set",
							Destination: &newStatus,
						},
						cli.StringFlag{
							Name:        "from-state, f",
							Usage:       "The state range start",
							Destination: &fromState,
						},
						cli.BoolFlag{
							Name:  "from-included, fi",
							Usage: "Include the state range start",
						},
						cli.StringFlag{
							Name:        "to-state, t",
							Usage:       "The state range end",
							Destination: &toState,
						},
						cli.BoolFlag{
							Name:  "to-included, ti",
							Usage: "Include the state range end",
						},
					},
					Action: setStatesStatuses,
				},
				{
					Name:    "merge",
					Aliases: []string{"n"},
					Hidden:  true,
					Usage:   "merge a states file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "states, s",
							Usage:       "States file path to merge",
							Destination: &statesPath,
						},
					},
					Action: mergeStates,
				},
			},
		},
	}

	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
