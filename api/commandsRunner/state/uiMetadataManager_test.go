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
	"strconv"
	"testing"

	"github.com/olebedev/config"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func TestGetUIConfigExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//	global.SetExtensionResourcePath("../../test/resource/extensions")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIConfigExtentionTest", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	configuration, err := GetUIMetaDataConfig("ext-template", "test-ui", []string{global.DefaultLanguage})
	if err != nil {
		t.Error(err.Error())
	}
	_, err = config.ParseYaml(string(configuration))
	if err != nil {
		t.Error(err.Error())
	}

	t.Log(string(configuration))
	global.RemoveTemp("TestGetUIConfigExtentionTest")
}

func TestGetUIMetadataParseConfigsExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIMetadataParseConfigsExtentionTest", "../../test/resource/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	cfg, err := getUIMetadataParseConfigs("ext-template", []string{global.DefaultLanguage})
	if err != nil {
		t.Error(err.Error())
	}
	configs, err := cfg.Map("")
	if err != nil {
		t.Error(err.Error())
	}
	n := len(configs)
	if n != 2 {
		t.Error("Expected to find 2 configs but found " + strconv.Itoa(n))
	}
	t.Log(strconv.Itoa(n))
	out, err := config.RenderYaml(cfg.Root)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(string(out))
	global.RemoveTemp("TestGetUIMetadataParseConfigsExtentionTest")
}

func TestGetUIConfigError(t *testing.T) {
	_, err := GetUIMetaDataConfig("does-not-exist", "test-ui", []string{global.DefaultLanguage})
	if err == nil {
		t.Error("An error should be raised as this file doesn't exists")
	}
}
