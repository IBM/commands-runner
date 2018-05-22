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
	"net/url"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/statusManager"
)

//GetCMStatus returns the configManager status
func (cmc *ConfigManagerClient) GetCMStatus() (string, error) {
	url := "status"
	data, errCode, err := cmc.restCall(http.MethodGet, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get status, please check logs")
	}
	//Convert the config to object
	var statuses statusManager.Statuses
	jsonErr := json.Unmarshal([]byte(data), &statuses)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		return "", jsonErr
	}
	//Generate the text format otherwize return the json
	if cmc.OutputFormat == "text" {
		out := ""
		for _, status := range statuses {
			out += fmt.Sprintf("%s: %s\n", status.Name, status.Status)
		}
		return out, nil
	}
	return cmc.convertJSONOrYAML(data)
}

//SetCMStatus set a status
func (cmc *ConfigManagerClient) SetCMStatus(name string, status string) (string, error) {
	url := "status?name=" + url.QueryEscape(name) + "&status=" + url.QueryEscape(status)
	_, errCode, err := cmc.restCall(http.MethodPut, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to set status, please check logs")
	}
	//Convert the config to object
	return cmc.GetCMStatus()
}
