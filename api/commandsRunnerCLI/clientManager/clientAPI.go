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

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//SetAPISetup set the API configuration
func SetAPISetup(urlIn string, outputFormat string, timeout string, caCertPath string, insecureSSL bool, token string) error {
	if urlIn == "" {
		urlIn = global.DefaultUrl
	}
	if outputFormat == "" {
		outputFormat = global.DefaultOutputFormat
	}
	if timeout == "" {
		timeout = strconv.Itoa(global.DefaultTimeout)
	}
	//Convert timeout to integer
	timeoutI, errInt := strconv.Atoi(timeout)
	if errInt != nil {
		return errInt
	}
	finalCACertPath, err := filepath.Abs(caCertPath)
	if err != nil {
		finalCACertPath = caCertPath
	}
	//Create the client object
	c := &CommandsRunnerClient{
		URL:          urlIn,
		OutputFormat: outputFormat,
		Timeout:      timeoutI,
		CACertPath:   finalCACertPath,
		InsecureSSL:  insecureSSL,
		Token:        token,
	}
	//Convert it as json
	data, err := json.MarshalIndent(c, "", "  ")
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
	client, err := NewClient("", "", "", "", u.Scheme != "https", "")
	_, errStatus := client.GetCMStatus()
	if errStatus != nil {
		return errors.New("wrong url, certificate or token or API server not ready yet")
	}
	return nil
}

//GetAPISetup retrieves the API Setup
func GetAPISetup(outputFormat string) (string, error) {
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
