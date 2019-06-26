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
package clientManager

import (
	"errors"
	"net/http"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

//GetUIConfig reeturns the uiconfig
func (crc *CommandsRunnerClient) GetTemplate(extensionName string, uiConfigName string) (string, error) {
	if extensionName == "" {
		extensionName = crc.DefaultExtensionName
	}
	url := "template?extension-name=" + extensionName + "&ui-metadata-name=" + uiConfigName
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get the ui metadata: " + data + ", please check logs for more details")
	}
	return data, nil
}
