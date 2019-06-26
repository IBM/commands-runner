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

import "github.com/olebedev/config"

//Convert JsonToYaml
func convertJSONToYAML(data string) (string, error) {
	cfg, err := config.ParseJson(data)
	if err != nil {
		return "", err
	}
	out, err := config.RenderYaml(cfg.Root)
	if err != nil {
		return "", err
	}
	return out, nil
}

//Convert to yaml only if the requested format is not json.
func (crc *CommandsRunnerClient) convertJSONOrYAML(data string) (string, error) {
	if crc.OutputFormat == "json" {
		return data, nil
	}
	return convertJSONToYAML(data)
}
