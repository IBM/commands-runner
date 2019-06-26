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
package global

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestExtractKey(t *testing.T) {
	t.Log("Entering... TextExtractKey")
	input, err := ExtractKey("../../test/resource/manifest/extension-manifest.yml", "states")
	var inYaml map[string]interface{}
	inYaml = make(map[string]interface{}, 0)
	err = yaml.Unmarshal(input, &inYaml)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(inYaml)
	if len(inYaml) > 1 {
		t.Error("out file contains more than 1 entry")
	}
	_, ok := inYaml["states"]
	if !ok {
		t.Error("states key not found")
	}
}

// func TestCopyTemp(t *testing.T) {
// 	log.SetLevel(log.DebugLevel)
// 	caller, err := CopyToTemp("TestCopyTemp", "../../test/data/extensions")
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	RemoveTemp("TestCopyTemp")
// 	caller, err = CopyToTemp("TestCopyTemp/test", "../../test/data/extensions/test-extensions.yml")
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	RemoveTemp("TestCopyTemp/test")
// 	t.Error("Caller:" + caller)
// }
