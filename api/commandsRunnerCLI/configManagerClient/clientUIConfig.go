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
	"net/http"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetUIConfig reeturns the uiconfig
func (cmc *ConfigManagerClient) GetUIConfig(uiConfig string) (string, error) {
	url := "uiconfig/" + uiConfig
	data, errCode, err := cmc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get uiconfig" + data + ", please check logs for more details")
	}
	return cmc.convertJSONOrYAML(data)
}
