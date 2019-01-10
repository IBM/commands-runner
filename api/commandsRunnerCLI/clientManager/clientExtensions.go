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
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
)

func (crc *CommandsRunnerClient) getExtensions(extensionsToList string, catalog bool) (string, error) {
	url := "/extensions?filter=" + extensionsToList + "&catalog=" + strconv.FormatBool(catalog)
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get extensions: " + data + ", please check log for more information")
	}
	return data, err
}

func (crc *CommandsRunnerClient) GetExtensions(extensionToList string, catalog bool) (string, error) {
	data, err := crc.getExtensions(extensionToList, catalog)
	if err != nil {
		return "", err
	}
	out := ""
	if crc.OutputFormat == "text" {
		var extensions state.Extensions
		jsonErr := json.Unmarshal([]byte(data), &extensions)
		if jsonErr != nil {
			return "", jsonErr
		}
		for key, v := range extensions.Extensions {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("name : %s\n", key)
			out += fmt.Sprintf("type : %s\n", v.Type)
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}
