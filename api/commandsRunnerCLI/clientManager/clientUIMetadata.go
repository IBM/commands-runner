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
	"errors"
	"net/http"
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetUIConfig reeturns the uiconfig
func (crc *CommandsRunnerClient) GetUIMetadata(extensionName string, uiConfigName string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	url := "uimetadata?extension-name=" + extensionName + "&ui-metadata-name=" + uiConfigName
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get the ui metadata: " + data + ", please check logs for more details")
	}
	return crc.convertJSONOrYAML(data)
}

func (crc *CommandsRunnerClient) GetUIMetadatas(extensionName string, namesOnly bool) (string, error) {
	url := "uimetadatas?names-only=" + strconv.FormatBool(namesOnly)
	if extensionName != "" {
		url += "&extension-name=" + extensionName
	}
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get the ui metadata: " + data + ", please check logs for more details")
	}
	return crc.convertJSONOrYAML(data)
}
