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
	"io"
	"net/http"
	"os"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
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
