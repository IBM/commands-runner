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

func TestGetUIConfigsExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionsPath("../../test/resource/extensions/")
	configuration, err := GetUIMetaDataConfigs("ext-template", false)
	if err != nil {
		t.Error(err.Error())
	}
	cfg, err := config.ParseJson(string(configuration))
	if err != nil {
		t.Error(err.Error())
	}
	configs, err := cfg.Map("ui_metadata")
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
