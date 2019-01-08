/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package state

import (
	"path/filepath"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
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

func TestIsExtension(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	extensionPath, err := global.CopyToTemp("TestIsExtension", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	isExtension, err := IsExtension("ext-template")
	if err != nil {
		t.Error(err.Error())
	}
	if !isExtension {
		t.Error("ext-template is an IBM extension")
	}
	global.RemoveTemp("TestIsExtension")
}

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
