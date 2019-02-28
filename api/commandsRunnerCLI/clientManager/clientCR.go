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
package clientManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetLogLevel of PCM
func (crc *CommandsRunnerClient) GetCRLogLevel() (string, error) {
	url := "cr/log/level"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get pcm log-level: " + data + ", please check log for more information")
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		cfg, err := config.ParseJson(data)
		if err != nil {
			return "", err
		}
		level, err := cfg.String("level")
		if err != nil {
			return "", err
		}
		out := fmt.Sprintf("level: %s\n", level)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

func (crc *CommandsRunnerClient) SetCRLogLevel(level string) (string, error) {
	url := "cr/log/level?level=" + level
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to set cr log level: " + data + ", please check log for more information")
	}
	return data, nil
}

//GetCRLogMaxBackups of CR
func (crc *CommandsRunnerClient) GetCRLogMaxBackups() (string, error) {
	url := "cr/log/max-backups"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get cr max-backups: " + data + ", please check log for more information")
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		cfg, err := config.ParseJson(data)
		if err != nil {
			return "", err
		}
		maxBackups, err := cfg.String("max_backups")
		if err != nil {
			return "", err
		}
		out := fmt.Sprintf("max-backups: %s\n", maxBackups)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

func (crc *CommandsRunnerClient) SetCRLogMaxBackups(maxBackups string) (string, error) {
	url := "cr/log/max-backups?max-backups=" + maxBackups
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to set cr log max-backups: " + data + ", please check log for more information")
	}
	return data, nil
}

//GetCRSettings of CR
func (crc *CommandsRunnerClient) GetCRSettings() (string, error) {
	url := "cr/settings"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get pcm log-level: " + data + ", please check log for more information")
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		var settings commandsRunner.Settings
		err := json.Unmarshal([]byte(data), &settings)
		if err != nil {
			return "", err
		}
		out := fmt.Sprintf("DeploymentName:       %s\n", settings.DeploymentName)
		out += fmt.Sprintf("defaultExtensionName: %s\n", settings.DefaultExtensionName)
		out += fmt.Sprintf("configRootKey       : %s\n", settings.ConfigRootKey)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//GetAbout of CR
func (crc *CommandsRunnerClient) GetCRAbout() (string, error) {
	url := "cr/about"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get about: " + data + ", please check log for more information")
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		cfg, err := config.ParseJson(data)
		if err != nil {
			return "", err
		}
		about, err := cfg.String("about")
		if err != nil {
			return "", err
		}
		out := fmt.Sprintf("About: %s\n", about)
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}
