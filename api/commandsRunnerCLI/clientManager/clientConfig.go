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
package clientManager

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/olebedev/config"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetConfig returns the config
func (crc *CommandsRunnerClient) GetConfig(extensionName string) (string, error) {
	url := "config"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest API
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get config: " + data + ",please check logs")
	}
	//Convert to text otherwize return the json
	if crc.OutputFormat == "text" {
		//	var configAux config.Config
		cfg, jsonErr := config.ParseJson(data)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		ps, jsonErr := cfg.Map(global.ConfigRootKey)
		//		jsonErr = json.Unmarshal([]byte(data), &configAux)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		out := ""
		for k, v := range ps {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("Name      : %s\n", k)
			out += fmt.Sprintf("Value     : %s\n", v)
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

func (crc *CommandsRunnerClient) GetProperty(extensionName string, propertyName string) (string, error) {
	url := "config/" + propertyName
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	//Call the rest API
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get config: " + data + ",please check logs")
	}
	//Convert to text otherwize return the json
	if crc.OutputFormat == "text" {
		//	var configAux config.Config
		cfg, jsonErr := config.ParseJson(data)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		name, jsonErr := cfg.String("name")
		//		jsonErr = json.Unmarshal([]byte(data), &configAux)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		value, jsonErr := cfg.Get("value")
		//		jsonErr = json.Unmarshal([]byte(data), &configAux)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		valueJson, jsonErr := config.RenderJson(value)
		if jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return "", jsonErr
		}
		out := ""
		out += fmt.Sprintf("Name      : %s\n", name)
		out += fmt.Sprintf("Value     : %s\n", valueJson)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//SetConfig saves config
func (crc *CommandsRunnerClient) SetConfig(extensionName string, configPath string) (string, error) {
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
	data, errCode, err := crc.RestCall(http.MethodPost, global.BaseURL, url, file, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to save the configuration: " + data + ", please check log for more information")
	}
	return "", nil
}
