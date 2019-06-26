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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func TestExtensionRegistered(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("Entering... check to see if an extension is registerd")
	extensionPath, err := global.CopyToTemp("TestExtensionRegistered", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	registered := IsExtensionRegistered("ext-template")
	if registered == false {
		t.Error("extension does not exist")
	}
	global.RemoveTemp("TestExtensionRegistered")
}

func TestIsIBMExtension(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestIsIBMExtension", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	isExtension, err := IsEmbeddedExtension("ext-template")
	if err != nil {
		t.Error(err.Error())
	}
	if !isExtension {
		t.Error("ext-template is an IBM extension")
	}
	global.RemoveTemp("TestIsIBMExtension")
}

// func TestIsExtension(t *testing.T) {
// 	//log.SetLevel(log.DebugLevel)
// 	extensionPath, err := global.CopyToTemp("TestIsExtension", "../../test/data/extensions/")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	SetExtensionsPath(extensionPath)
// 	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
// 	isExtension, err := IsExtension("ext-template")
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	if !isExtension {
// 		t.Error("ext-template is an IBM extension")
// 	}
// 	global.RemoveTemp("TestIsExtension")
// }

func TestListRegisteredExtensions(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestListRegisteredExtensions", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	extensions, err := ListExtensions("", false)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("%v", extensions)
	global.RemoveTemp("TestListRegisteredExtensions")
}

func TestExtensionPathWithVersion(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestExtensionPathWithVersion", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	path, err := getEmbeddedExtensionRepoPath("ext-template-v")
	if err != nil {
		t.Error(err.Error())
	}
	expectedPath := filepath.Join(embeddedExtensionsRepositoryPath, "ext-template-v", "1.0.0")
	if expectedPath != path {
		t.Errorf("Got %s expected %s", path, expectedPath)
	}
	global.RemoveTemp("TestExtensionPathWithVersion")
}

func TestExtensionPathWithOutVersion(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestExtensionPathWithOutVersion", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	path, err := getEmbeddedExtensionRepoPath("ext-template")
	if err != nil {
		t.Error(err.Error())
	}
	expectedPath := filepath.Join(embeddedExtensionsRepositoryPath, "ext-template")
	if expectedPath != path {
		t.Errorf("Got %s expected %s", path, expectedPath)
	}
	global.RemoveTemp("TestExtensionPathWithOutVersion")
}

func TestExtensionPathForNonExistExtension(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestExtensionPathForNonExistExtension", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	_, err = getEmbeddedExtensionRepoPath("not-exist")
	if err == nil {
		t.Error("Expecting an error as extension name not-exist doesn't exist")
	}
	global.RemoveTemp("TestExtensionPathForNonExistExtension")
}

func TestParseIBMExtension(t *testing.T) {
	var extensionList Extensions
	resource, err := ioutil.ReadFile("../../test/resource/ibm-extensions.yml")
	if err != nil {
		t.Error(err.Error())
	}
	err = yaml.Unmarshal(resource, &extensionList)
	if err != nil {
		t.Error(err.Error())
	}

}
