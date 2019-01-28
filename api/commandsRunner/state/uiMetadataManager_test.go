/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package state

import (
	"strconv"
	"testing"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
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
