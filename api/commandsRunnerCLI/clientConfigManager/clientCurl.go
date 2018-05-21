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
	"io"
	"net/http"
	"os"
)

//GetBOM returns the BOM
func (cmc *ConfigManagerClient) Curl(method string, url string, dataPath string) (string, error) {
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
	data, errCode, err := cmc.restCall(method, url, file, nil)
	if err != nil {
		return data, err
	}
	if errCode != http.StatusOK {
		return data, errors.New("Unable to get bom, please check logs")
	}
	return cmc.convertJSONOrYAML(data)
}
