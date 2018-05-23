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
package configManagerClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/configManager"
)

//GetConfig returns the config
func (cmc *ConfigManagerClient) GetConfig(extensionName string) (string, error) {
	url := "config"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest API
	data, errCode, err := cmc.restCall(http.MethodGet, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get config:" + data + ",please check logs")
	}
	//Convert to text otherwize return the json
	if cmc.OutputFormat == "text" {
		var config configManager.Config
		jsonErr := json.Unmarshal([]byte(data), &config)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		out := ""
		for k, v := range config.Properties {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("Name      : %s\n", k)
			out += fmt.Sprintf("Value     : %s\n", v)
		}
		return out, nil
	}
	return cmc.convertJSONOrYAML(data)
}

//SetConfig saves config
func (cmc *ConfigManagerClient) SetConfig(extensionName string, configPath string) (string, error) {
	if configPath == "" {
		errConfigPath := errors.New("config file missing")
		return "", errConfigPath
	}
	url := "config"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest API
	file, errFile := os.Open(configPath)
	if errFile != nil {
		return "", errFile
	}
	data, errCode, err := cmc.restCall(http.MethodPost, url, file, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to save the configuration:" + data + ", please check log for more information")
	}
	return "", nil
}
