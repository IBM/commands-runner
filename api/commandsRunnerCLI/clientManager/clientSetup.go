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
//Package clientManager provides a CLI to end-users
package clientManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

//SetAPISetup set the API configuration
func SetClientSetup(urlIn string, outputFormat string, timeout string, caCertPath string, insecureSSL string, token string, defaultExtensionName string) error {
	var clientManager CommandsRunnerClient
	//Read existing
	data, errFile := ioutil.ReadFile(configFilePath)
	if errFile == nil {
		//Convert the config to object
		jsonErr := json.Unmarshal([]byte(data), &clientManager)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return jsonErr
		}
	} else {
		//create new
		if urlIn == "" {
			urlIn = global.DefaultUrl
		}
		if outputFormat == "" {
			outputFormat = global.DefaultOutputFormat
		}
		//Create the client object
		c := &CommandsRunnerClient{
			URL:                  urlIn,
			OutputFormat:         outputFormat,
			Timeout:              global.DefaultTimeout,
			CACertPath:           "",
			InsecureSSL:          true,
			Token:                "",
			DefaultExtensionName: "",
		}
		clientManager = *c
	}
	//Calculate values
	if timeout == "" {
		timeout = strconv.Itoa(global.DefaultTimeout)
	}
	//Convert timeout to integer
	timeoutI, errInt := strconv.Atoi(timeout)
	if errInt != nil {
		return errInt
	}
	var finalCACertPath string
	var err error
	if caCertPath != "" {
		finalCACertPath, err = filepath.Abs(caCertPath)
		if err != nil {
			finalCACertPath = caCertPath
		}
	}
	var insecureSSLBool bool
	if insecureSSL != "" {
		insecureSSLBool, err = strconv.ParseBool(insecureSSL)
		if err != nil {
			return err
		}
	}
	//Set values
	if urlIn != "" {
		clientManager.URL = urlIn
	}
	if outputFormat != "" {
		clientManager.OutputFormat = outputFormat
	}
	if timeout != "" {
		clientManager.Timeout = timeoutI
	}
	if caCertPath != "" {
		clientManager.CACertPath = finalCACertPath
	}
	if insecureSSL != "" {
		clientManager.InsecureSSL = insecureSSLBool
	}
	if token != "" {
		clientManager.Token = token
	}
	if defaultExtensionName != "" {
		clientManager.DefaultExtensionName = defaultExtensionName
	}
	//Convert it as json
	data, err = json.MarshalIndent(clientManager, "", "  ")
	if err != nil {
		return err
	}
	//Write the config
	errWrite := ioutil.WriteFile(configFilePath, data, 0644)
	if errWrite != nil {
		fmt.Print("Failed to create config file:" + errWrite.Error())
		return errWrite
	}
	//Test the config
	u, err := url.Parse(urlIn)
	if err != nil {
		return err
	}
	insecureSSL = strconv.FormatBool(u.Scheme != "https")
	client, err := NewClient("", "", "", "", insecureSSL, "", "")
	_, errStatus := client.GetCMStatus()
	if errStatus != nil {
		return errors.New("wrong url, certificate or token or API server not ready yet:" + errStatus.Error())
	}
	return nil
}

//GetAPISetup retrieves the API Setup
func GetClientSetup(outputFormat string) (string, error) {
	//Read the config file
	data, errFile := ioutil.ReadFile(configFilePath)
	if errFile != nil {
		fmt.Print(errFile.Error())
		return "", errFile
	}
	//Convert the config to object
	var clientManager CommandsRunnerClient
	jsonErr := json.Unmarshal([]byte(data), &clientManager)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		return "", jsonErr
	}
	//Overwrite the format with the requested format
	if outputFormat != "" {
		clientManager.OutputFormat = outputFormat
	}
	//Generate the text format otherwize return the json
	switch clientManager.OutputFormat {
	case "text":
		out := fmt.Sprintf("url:     %s\n", clientManager.URL)
		out += fmt.Sprintf("Format:  %s\n", clientManager.OutputFormat)
		out += fmt.Sprintf("Timeout: %d\n", clientManager.Timeout)
		out += fmt.Sprintf("CACertPath: %s\n", clientManager.CACertPath)
		out += fmt.Sprintf("InsecureSSL: %t\n", clientManager.InsecureSSL)
		out += fmt.Sprintf("Token: %s\n", clientManager.Token)
		out += fmt.Sprintf("DefaultExtensionName: %s\n", clientManager.DefaultExtensionName)
		return out, nil
	case "json":
		return string(data), nil
	case "yaml":
		return convertJSONToYAML(string(data))
	default:
		return "", errors.New("Format " + clientManager.OutputFormat + " not supported")
	}

}

//RemoveAPISetup removes the file where the API Setup is stored
func RemoveAPISetup() error {
	err := os.Remove(configFilePath)
	return err
}
