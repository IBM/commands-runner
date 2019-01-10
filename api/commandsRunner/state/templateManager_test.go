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
	"bytes"
	"fmt"
	"testing"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func TestTraverseProperties(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	var properties []interface{}
	properties = make([]interface{}, 0)
	p1 := make(map[string]interface{}, 0)
	p1["name"] = "param1"
	p1["description"] = "Description 1"
	p1["sample_value"] = "Eg: sample_value 1"
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
	expectedOut := `# Description 1
param1: "Eg: sample_value 1"
# Description 
# 2
param2: "Eg: sample_value 2"
# # Description 
# 3
# param3:
#   # Description 31
#   param31: "Eg: 
# sample_value 31"
# Description 4
param4: "Eg: sample_value 4"
`
	if out.String() != expectedOut {
		t.Errorf("expecting: \n%s \ngot \n%s", expectedOut, out.String())
	}
	t.Logf("\n%s", out.String())
}

func TestGetUIMetadataTemplate(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionPath, err := global.CopyToTemp("TestGetUIMetadataTemplate", "../../test/data/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	SetExtensionsPath(extensionPath)
	data, err := getUIMetadataTemplate("ext-template", "test-ui")
	if err != nil {
		t.Logf("\n%s", data)
		t.Error(err.Error())
	}
	t.Logf("\n%s", data)
	global.RemoveTemp("TestGetUIMetadataTemplate")
}
