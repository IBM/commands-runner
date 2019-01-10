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
	"net/url"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/status"
)

//GetCMStatus returns the configManager status
func (crc *CommandsRunnerClient) GetCMStatus() (string, error) {
	url := "status"
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get status: " + data + ", please check log for more information")
	}
	//Convert the config to object
	var statuses status.Statuses
	jsonErr := json.Unmarshal([]byte(data), &statuses)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		return "", jsonErr
	}
	//Generate the text format otherwize return the json
	if crc.OutputFormat == "text" {
		out := ""
		for _, status := range statuses {
			out += fmt.Sprintf("%s: %s\n", status.Name, status.Status)
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}

//SetCMStatus set a status
func (crc *CommandsRunnerClient) SetCMStatus(name string, status string) (string, error) {
	url := "status?name=" + url.QueryEscape(name) + "&status=" + url.QueryEscape(status)
	data, errCode, err := crc.RestCall(http.MethodPut, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to set status: " + data + ", please check log for more information")
	}
	//Convert the config to object
	return crc.GetCMStatus()
}
