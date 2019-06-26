/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
*/
package helloWorld

import (
	"net/http"

	"github.com/IBM/commands-runner/api/commandsRunnerCLI/clientManager"
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
