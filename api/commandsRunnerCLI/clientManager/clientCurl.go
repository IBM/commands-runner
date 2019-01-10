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
	"io"
	"net/http"
	"os"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//GetBOM returns the BOM
func (crc *CommandsRunnerClient) Curl(method string, url string, dataPath string) (string, error) {
	if method == "" {
		method = "GET"
	}
	var file io.Reader
	if dataPath != "" {
		fileOS, errFile := os.Open(dataPath)
		if errFile != nil {
			return "", errFile
		}
		file = fileOS
	}
	data, errCode, err := crc.RestCall(method, global.BaseURL, url, file, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get bom: " + data + ", please check log for more information")
	}
	return crc.convertJSONOrYAML(data)
}
