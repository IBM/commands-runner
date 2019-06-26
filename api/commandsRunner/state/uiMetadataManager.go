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
package state

import (
	"errors"

	"github.com/IBM/commands-runner/api/commandsRunner/global"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
)

func GetUIMetaDataConfig(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... GetUIMetaDataConfig")
	log.Debugf("extensionName=%s", extensionName)
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	if uiMetadataName == "" {
		uiMetadataName = global.DefaultUIMetaDataName
	}
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	raw, e := getUIMetadataConfig(extensionName, uiMetadataName, langs)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadataConfig(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataConfig")
	cfg, err := getUIMetadataParseConfig(extensionName, uiMetadataName, langs)
	if err == nil {
		uiConfigFilefg, err := config.ParseYaml("ui_metadata:")
		if err != nil {
			return nil, err
		}
		err = uiConfigFilefg.Set("ui_metadata", cfg.Root)
		if err != nil {
			return nil, err
		}
		out, err := config.RenderJson(uiConfigFilefg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available for " + extensionName + " and " + uiMetadataName)
}

func getUIMetadataParseConfig(extensionName string, uiMetadataName string, langs []string) (cfg *config.Config, err error) {
	log.Debug("Entering in... getUIMetadataParseConfig")
	cfg, err = getUIMetadataParseConfigs(extensionName, langs)
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get(uiMetadataName)
	return cfg, err
}
