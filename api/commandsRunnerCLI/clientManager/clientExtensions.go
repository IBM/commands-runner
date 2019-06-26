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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
)

func (crc *CommandsRunnerClient) getExtensions(extensionsToList string, catalog bool) (string, error) {
	url := "/extensions?filter=" + extensionsToList + "&catalog=" + strconv.FormatBool(catalog)
	data, errCode, err := crc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get extensions: " + data + ", please check log for more information")
	}
	return data, err
}

func (crc *CommandsRunnerClient) GetExtensions(extensionToList string, catalog bool) (string, error) {
	data, err := crc.getExtensions(extensionToList, catalog)
	if err != nil {
		return "", err
	}
	out := ""
	if crc.OutputFormat == "text" {
		var extensions state.Extensions
		jsonErr := json.Unmarshal([]byte(data), &extensions)
		if jsonErr != nil {
			return "", jsonErr
		}
		for key, v := range extensions.Extensions {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("name : %s\n", key)
			out += fmt.Sprintf("type : %s\n", v.Type)
		}
		return out, nil
	}
	return crc.convertJSONOrYAML(data)
}
