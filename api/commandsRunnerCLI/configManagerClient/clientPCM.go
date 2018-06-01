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
	"errors"
	"fmt"
	"net/http"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetLogLevel of PCM
func (cmc *ConfigManagerClient) GetPCMLogLevel() (string, error) {
	url := "pcm/log/level"
	data, errCode, err := cmc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get pcm log-level, please check logs")
	}
	//Generate the text format otherwize return the json
	if cmc.OutputFormat == "text" {
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
	return cmc.convertJSONOrYAML(data)
}

func (cmc *ConfigManagerClient) SetPCMLogLevel(level string) (string, error) {
	url := "pcm/log/level?level=" + level
	data, errCode, err := cmc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to set pcm log level, please check logs")
	}
	return data, nil
}
