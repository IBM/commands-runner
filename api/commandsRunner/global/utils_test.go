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
