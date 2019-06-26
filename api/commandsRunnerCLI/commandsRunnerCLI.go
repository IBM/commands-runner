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
package commandsRunnerCLI

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/IBM/commands-runner/api/commandsRunnerCLI/clientManager"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

var URL string
var OutputFormat string
var Timeout string
var CACertPath string
var InsecureSSL string
var Token string
var DefaultExtensionName string

func Client() *cli.App {
	var state string
	var newStatus string
	var stateTimeout string
	var configPath string
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
	var logMaxBackups string

	var statusName string
	var status string

	var uiMetadataName string

	var propertyName string

	getStatus := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		data, errClient := clientManager.GetClientSetup(OutputFormat)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		fmt.Println(data)
		return nil
	}

	setClientSetup := func(c *cli.Context) error {
		errClient := clientManager.SetClientSetup(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		return nil
	}

	removeAPI := func(c *cli.Context) error {
		errClient := clientManager.RemoveAPISetup()
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		return nil
	}

	createToken := func(c *cli.Context) error {
		if tokenOutputFilePath != "" {
			return clientManager.NewTokenFile(tokenOutputFilePath)
		}
		data := clientManager.NewToken()
		fmt.Print(data)
		return nil
	}

	getLogs := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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

	getConfig := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data := ""
		var err error
		if propertyName == "" {
			data, err = client.GetConfig(extensionName)
		} else {
			data, err = client.GetProperty(extensionName, propertyName)
		}
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	setConfig := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetConfig(extensionName, configPath)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	generateConfig := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GenerateConfig(extensionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	validateConfig := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.ValidateConfig(extensionName)
		fmt.Print(data)
		//Either 200 (ok), 299 (Warning), 406 (Error) or real error
		if err != nil {
			code, errCode := strconv.Atoi(err.Error())
			//Can not convert
			if errCode != nil {
				fmt.Println(err.Error())
				return err
			}
			switch code {
			case http.StatusOK:
				return nil
			case http.StatusNotAcceptable:
				return err
			case 299:
				return nil
			}
			return err
		}
		return nil
	}

	deploy := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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

	setMock := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		_, err := client.SetMockEngine(c.Bool("mock"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	}

	getMock := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetMockEngine()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	getUIMetaData := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		var data string
		if c.Bool("all") {
			data, errClient = client.GetUIMetadatas(extensionName, c.Bool("names-only"))
		} else {
			data, errClient = client.GetUIMetadata(extensionName, uiMetadataName)
		}
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		fmt.Println(data)
		return nil
	}

	getTemplate := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, errClient := client.GetTemplate(extensionName, uiMetadataName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		fmt.Println(data)
		return nil
	}

	setState := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetStates(extensionName, "", false, false)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	setStates := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.InsertStateStates(extensionName, statePosition, stateName, c.Bool("before"), statePath, insertExtensionName, c.Bool("overwrite"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	deleteStateStates := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetStates(extensionName, searchStatus, c.Bool("extension"), c.Bool("recursive"))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Print(data)
		return nil
	}

	register := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
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

	getCRLogLevel := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetCRLogLevel()
		fmt.Print(data)
		return err
	}

	getCRLogMaxBackups := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetCRLogMaxBackups()
		fmt.Print(data)
		return err
	}

	getCRAbout := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetCRAbout()
		fmt.Print(data)
		return err
	}

	setCRLogLevel := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetCRLogLevel(logLevel)
		fmt.Print(data)
		return err
	}

	setCRLogMaxBackups := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.SetCRLogMaxBackups(logMaxBackups)
		fmt.Print(data)
		return err
	}

	getCRSettings := func(c *cli.Context) error {
		client, errClient := clientManager.NewClient(URL, OutputFormat, Timeout, CACertPath, InsecureSSL, Token, DefaultExtensionName)
		if errClient != nil {
			fmt.Println(errClient.Error())
			return errClient
		}
		data, err := client.GetCRSettings()
		fmt.Print(data)
		return err
	}

	app := cli.NewApp()
	app.Usage = "Config Manager for Cloud Foundry installation"
	app.Description = "CLI to manage initial Cloud Foundry installation"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, u",
			Usage:       "API Url",
			Destination: &URL,
		},
		cli.StringFlag{
			Name:        "format, f",
			Usage:       "Output format (text, json, yaml)",
			Destination: &OutputFormat,
		},
		cli.StringFlag{
			Name:        "timeout, t",
			Usage:       "RestAPI timeout in seconds",
			Destination: &Timeout,
		},
		cli.StringFlag{
			Name:        "cacert",
			Usage:       "CA Cert path",
			Destination: &CACertPath,
		},
		cli.StringFlag{
			Name:        "insecure, s",
			Usage:       "true/false, false turn off SSL verification",
			Destination: &InsecureSSL,
		},
		cli.StringFlag{
			Name:        "token",
			Usage:       "Token",
			Destination: &Token,
		},
		cli.StringFlag{
			Name:        "default-extension-name, e",
			Usage:       "Default extension name",
			Destination: &DefaultExtensionName,
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
					Action:  setClientSetup,
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
					Usage:       "Which extensions to list (embedded or custom), leave empty for all",
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
			Action: getConfig,
			Subcommands: []cli.Command{
				{
					Name:  "uimetadata",
					Usage: "Get the ui metadata for extension and configuration name",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "extension, e",
							Usage:       "Extension name",
							Destination: &extensionName,
						},
						cli.StringFlag{
							Name:        "config, c",
							Usage:       "Configuration name (default is 'default'",
							Destination: &uiMetadataName,
						},
						cli.BoolFlag{
							Name:  "all, a",
							Usage: "Get all ui metadata configuration of a given extension or all extensions if none specified",
						},
						cli.BoolFlag{
							Name:  "names-only, n",
							Usage: "Get only the configuration names",
						},
					},
					Action: getUIMetaData,
				},
				{
					Name:  "template",
					Usage: "Get the template for extension and configuration name",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "extension, e",
							Usage:       "Extension name",
							Destination: &extensionName,
						},
						cli.StringFlag{
							Name:        "config, c",
							Usage:       "Configuration name (default is 'default'",
							Destination: &uiMetadataName,
						},
					},
					Action: getTemplate,
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
					Name:    "save",
					Aliases: []string{"s"},
					Usage:   "Save extension configuration",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "config, c",
							Usage:       "Configuration file",
							Destination: &configPath,
						},
					},
					Action: setConfig,
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
		/*            CR                  */
		{
			Name:  "cr",
			Usage: "cr management",
			Subcommands: []cli.Command{
				{
					Name:   "about",
					Usage:  "About",
					Action: getCRAbout,
				},
				{
					Name:   "log-level",
					Usage:  "Get the current log level",
					Action: getCRLogLevel,
				},
				{
					Name:   "set-log-level",
					Usage:  "Set the log level",
					Action: setCRLogLevel,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "level, l",
							Usage:       "Requested log level",
							Destination: &logLevel,
						},
					},
				},
				{
					Name:   "log-max-backups",
					Usage:  "Get the current log max backups",
					Action: getCRLogMaxBackups,
				},
				{
					Name:   "set-log-max-backups",
					Usage:  "Set the log max backups",
					Action: setCRLogMaxBackups,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "max-backups, mb",
							Usage:       "Requested log max backups",
							Destination: &logMaxBackups,
						},
					},
				},
				{
					Name:   "settings",
					Usage:  "Get the current commands runner settings",
					Action: getCRSettings,
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
		/*            config                  */
		{
			Name:  "config",
			Usage: "Manage configuration (get, save)",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
				cli.StringFlag{
					Name:        "property-name, p",
					Usage:       "Property name",
					Destination: &propertyName,
				},
			},
			Action: getConfig,
			Subcommands: []cli.Command{
				{
					Name:    "save",
					Aliases: []string{"s"},
					Usage:   "Save the configuration",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "config, c",
							Usage:       "Configuration file",
							Destination: &configPath,
						},
					},
					Action: setConfig,
				},
				{
					Name:    "validate",
					Aliases: []string{"v"},
					Usage:   "Validate the configuration",
					Action:  validateConfig,
				},
				{
					Name:    "generate-config",
					Aliases: []string{"g"},
					Hidden:  true,
					Usage:   "Generate configuration based on the save configuration",
					Action:  generateConfig,
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
				{
					Name:    "set-mock",
					Aliases: []string{"sm"},
					Usage:   "Set mock mode for the engine",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "mock, m",
							Usage: "If set then the engine will skip all scripts",
						},
					},
					Action: setMock,
				},
				{
					Name:    "get-mock",
					Aliases: []string{"gm"},
					Usage:   "Get the mock mode if the engine",
					Action:  getMock,
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
			UsageText: "<client> state [-e <extension_name>] -s <state> set --status <new_status>",
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
		/*            UIMETADATA                  */
		{
			Name:  "uimetadata",
			Usage: "UI Metadata management",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
				cli.StringFlag{
					Name:        "config, c",
					Usage:       "Configuration name",
					Destination: &uiMetadataName,
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Get all ui metadata configuration of a given extension or all extensions if none specified",
				},
				cli.BoolFlag{
					Name:  "names-only, n",
					Usage: "Get only the configuration names",
				},
			},
			Action: getUIMetaData,
		},
		/*            TEMPLATE                  */
		{
			Name:  "template",
			Usage: "Template management",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension, e",
					Usage:       "Extension name",
					Destination: &extensionName,
				},
				cli.StringFlag{
					Name:        "config, c",
					Usage:       "Configuration name",
					Destination: &uiMetadataName,
				},
			},
			Action: getTemplate,
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
						cli.BoolFlag{
							Name:  "extension, e",
							Usage: "Search extension states only",
						},
						cli.BoolFlag{
							Name:  "recursive, r",
							Usage: "Search recursively in the extension states",
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
						cli.BoolFlag{
							Name:  "overwrite, o",
							Usage: "overwrite the state if already exists",
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
	return app
}
