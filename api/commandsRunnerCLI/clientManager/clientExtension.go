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
package clientManager

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func (crc *CommandsRunnerClient) RegisterExtension(pathToZip, extensionName string, force bool) (string, error) {
	httpCode, err := crc.registerExtension(pathToZip, extensionName, force)
	if err != nil {
		return "Could not create extension\n", err
	}

	if httpCode == http.StatusConflict {
		return extensionName + " already exists", nil
	}
	return "Created " + extensionName + "\n", nil
}

func (crc *CommandsRunnerClient) UnregisterExtension(extensionName string) (string, error) {
	err := crc.unregisterExtension(extensionName)
	if err != nil {
		return "Could not delete extension\n", err
	}
	return "Deleted " + extensionName + "\n", nil
}

func (crc *CommandsRunnerClient) registerExtension(pathToZip, extensionName string, force bool) (int, error) {
	url := "extension"
	if extensionName != "" {
		url += "?extension-name=" + extensionName
	}
	headers := make(map[string]string)

	if pathToZip != "" {
		headers["Content-Type"] = "application/zip"
		headers["Content-Disposition"] = "upload; filename=" + filepath.Base(pathToZip)
	}
	//	headers["Extension-Name"] = extensionName
	headers["Force"] = strconv.FormatBool(force)
	//	url += "&force=" + strconv.FormatBool(force)
	var file io.Reader
	var errFile error
	if pathToZip != "" {
		file, errFile = os.Open(pathToZip)
		if errFile != nil {
			return http.StatusInternalServerError, errFile
		}
	}
	body, httpCode, err := crc.RestCall(http.MethodPost, global.BaseURL, url, file, headers)
	if httpCode != http.StatusConflict && httpCode != http.StatusCreated && httpCode != http.StatusOK {
		if body != "" {
			return httpCode, errors.New(body)
		}
		return httpCode, errors.New("Unable to get extensions, please check logs")
	}
	return httpCode, err
}

func (crc *CommandsRunnerClient) unregisterExtension(extensionName string) error {
	if extensionName == "" {
		return errors.New("Extension name missing")
	}
	url := "extension?extension-name=" + extensionName

	response, errCode, err := crc.RestCall(http.MethodDelete, global.BaseURL, url, nil, nil)
	if err != nil {
		return errors.New(err.Error())
	}
	if errCode != http.StatusOK {
		return errors.New(response)
	}
	return err
}
