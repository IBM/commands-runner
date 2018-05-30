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
	"path/filepath"
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func (cmc *ConfigManagerClient) RegisterExtension(pathToZip, extensionName string, force bool) (string, error) {
	httpCode, err := cmc.registerExtension(pathToZip, extensionName, force)
	if err != nil {
		return "Could not create extension\n", err
	}

	if httpCode == http.StatusConflict {
		return extensionName + " already exists", nil
	}
	return "Created " + extensionName + "\n", nil
}

func (cmc *ConfigManagerClient) UnregisterExtension(extensionName string) (string, error) {
	err := cmc.unregisterExtension(extensionName)
	if err != nil {
		return "Could not delete extension\n", err
	}
	return "Deleted " + extensionName + "\n", nil
}

func (cmc *ConfigManagerClient) registerExtension(pathToZip, extensionName string, force bool) (int, error) {
	url := "extension"
	headers := make(map[string]string)

	if pathToZip != "" {
		headers["Content-Type"] = "application/zip"
		headers["Content-Disposition"] = "upload; filename=" + filepath.Base(pathToZip)
	}
	headers["Extension-Name"] = extensionName
	headers["Force"] = strconv.FormatBool(force)
	var file io.Reader
	var errFile error
	if pathToZip != "" {
		file, errFile = os.Open(pathToZip)
		if errFile != nil {
			return http.StatusInternalServerError, errFile
		}
	}
	body, httpCode, err := cmc.RestCall(http.MethodPost, global.BaseURL, url, file, headers)
	if httpCode != http.StatusConflict && httpCode != http.StatusCreated && httpCode != http.StatusOK {
		if body != "" {
			return httpCode, errors.New(body)
		}
		return httpCode, errors.New("Unable to get extensions, please check logs")
	}
	return httpCode, err
}

func (cmc *ConfigManagerClient) unregisterExtension(extensionName string) error {
	if extensionName == "" {
		return errors.New("Extension name missing")
	}
	url := "extension?name=" + extensionName

	response, errCode, err := cmc.RestCall(http.MethodDelete, global.BaseURL, url, nil, nil)
	if err != nil {
		return errors.New(err.Error())
	}
	if errCode != http.StatusOK {
		return errors.New(response)
	}
	return err
}
