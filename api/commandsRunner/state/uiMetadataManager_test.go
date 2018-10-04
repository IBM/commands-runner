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
	"strconv"
	"testing"

	"github.com/olebedev/config"
)

func TestGetUIConfigExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//	global.SetExtensionResourcePath("../../test/resource/extensions")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/resource/extensions/")
	configuration, err := GetUIMetaDataConfig("ext-template", "test-ui")
	if err != nil {
		t.Error(err.Error())
	}
	_, err = config.ParseYaml(string(configuration))
	if err != nil {
		t.Error(err.Error())
	}

	t.Log(string(configuration))
}

func TestGetUIMetadataParseConfigsExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionPath("../../test/resource/extensions/")
	cfg, err := getUIMetadataParseConfigs("ext-template")
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
}

func TestGetUIConfigError(t *testing.T) {
	_, err := GetUIMetaDataConfig("does-not-exist", "test-ui")
	if err == nil {
		t.Error("An error should be raised as this file doesn't exists")
	}
}
