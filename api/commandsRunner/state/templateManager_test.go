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
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"
	"github.com/olebedev/config"
)

func TestTraverseProperties(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	var properties []interface{}
	properties = make([]interface{}, 0)
	p1 := make(map[string]interface{}, 0)
	p1["name"] = "param1"
	p1["sample_value"] = "Eg: sample_value 1"
	p1["mandatory"] = true
	properties = append(properties, p1)
	p2 := make(map[string]interface{}, 0)
	p2["name"] = "param2"
	p2["description"] = "Description \n2"
	p2["sample_value"] = "Eg: sample_value 2"
	properties = append(properties, p2)
	p3 := make(map[string]interface{}, 0)
	p3["name"] = "param3"
	p3["description"] = "Description \n3"
	p3["mandatory"] = false
	p31 := make(map[string]interface{}, 0)
	p31["name"] = "param31"
	p31["description"] = "Description 31"
	p31["sample_value"] = "Eg: \nsample_value 31"
	var p3props []interface{}
	p3props = make([]interface{}, 0)
	p3props = append(p3props, p31)
	p3["properties"] = p3props
	properties = append(properties, p3)
	p4 := make(map[string]interface{}, 0)
	p4["name"] = "param4"
	p4["description"] = "Description 4"
	p4["sample_value"] = "Eg: sample_value 4"
	properties = append(properties, p4)
	fmt.Printf("%v", properties)
	var path string
	out := bytes.NewBufferString("")
	err := traverseProperties(properties, false, true, nil, printPropertyCallBack(), path, out)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(out.String())
	_, err = config.ParseYaml(out.String())
	if err != nil {
		t.Error(err.Error())
	}
	expectedOut := `# No description - sample_value: Eg: sample_value 1
param1: # Eg: sample_value 1
# Description 
# 2 - sample_value: Eg: sample_value 2
param2: # Eg: sample_value 2
# # Description 
# 3
# param3:
#   # Description 31 - sample_value: Eg: 
# sample_value 31
#   param31: # Eg: 
# sample_value 31
# Description 4 - sample_value: Eg: sample_value 4
param4: # Eg: sample_value 4
`
	if out.String() != expectedOut {
		t.Errorf("expecting: \n%s \ngot \n%s", expectedOut, out.String())
	}
	t.Logf("\n%s", out.String())
}

func TestGetUIMetadataTemplate(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionsPath, err := global.CopyToTemp("TestGetUIMetadataTemplate", "../../test/data/extensions")
	//As for test the extension are not really registered, we need to load the translation manually
	i18nUtils.LoadTranslationFilesFromDir(filepath.Join(extensionsPath, "embedded", "ext-template", i18nUtils.I18nDirectory))
	t.Log(extensionsPath)
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionsPath)
	data, err := getUIMetadataTemplate("ext-template", "test-ui", []string{global.DefaultLanguage})
	if err != nil {
		t.Logf("\n%s", data)
		t.Error(err.Error())
	}
	t.Logf("\n%s", data)
	global.RemoveTemp("TestGetUIMetadataTemplate")
}
