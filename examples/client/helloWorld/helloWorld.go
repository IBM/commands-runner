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
package helloWorld

import (
	"net/http"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunnerCLI/clientManager"
)

type MyCommandsRunnerClient struct {
	CMC *clientManager.CommandsRunnerClient
}

func NewClient(urlIn string, outputFormat string, timeout string, caCertPath string, insecureSSL string, token string, defaultExtensionName string) (*MyCommandsRunnerClient, error) {
	client, errClient := clientManager.NewClient(urlIn, outputFormat, timeout, caCertPath, insecureSSL, token, defaultExtensionName)
	if errClient != nil {
		return nil, errClient
	}
	myClient := &MyCommandsRunnerClient{client}
	return myClient, nil
}

func (crc *MyCommandsRunnerClient) HelloWorld() (string, error) {
	url := "myurl"
	data, _, err := crc.CMC.RestCall(http.MethodGet, "/", url, nil, nil)
	if err != nil {
		return "", err
	}
	return data, nil
}
