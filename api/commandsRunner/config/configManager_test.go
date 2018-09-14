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
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/properties"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
)

const COPYRIGHT_TEST string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

//var bmxConfigString string = "{\"Prop1\":{\"name\":\"Prop1\",\"value\":\"Val1\"},\"Prop2\":{\"name\":\"Prop2\",\"value\":\"Val2\"},\"subnet\":{\"name\":\"subnet\",\"value\":\"192.168.100.0/24\"}}"

var configString string = global.ConfigRootKey + ":\n  env_name: \"itdove\"\n  host_directory: \"/itdove/data\"\n  subnet: \"192.168.100.0/24\""

//var global.ConfigDirectory string = "../../test/resource"
//var properties properties.Properties

func TestSetConfigPath(t *testing.T) {
	t.Log("Entering... TestSetConfigPath")
	global.ConfigDirectory = "../../test/resource"
	os.Remove(global.ConfigDirectory)
	SetConfigPath(global.ConfigDirectory)
}

func TestSetProperties(t *testing.T) {
	t.Log("Entering... TestSetproperties.Properties")
	props = make(properties.Properties)
	global.ConfigDirectory = "../../test/resource"
	state.SetExtensionPath("../../test/resource/extensions")
	//	t.Error(global.ConfigDirectory)
	os.MkdirAll(global.ConfigDirectory, 0744)
	props["Prop3"] = "Val3"
	props["Prop4"] = "Val4"
	props["subnet"] = "192.168.100.0/24"
	err := SetProperties("config-manager-test", props)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetProperties(t *testing.T) {
	t.Log("Entering... TestGetproperties.Properties")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	state.SetExtensionPath("../../test/resource/extensions")
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	t.Log("dataDirectory:" + dataDirectory)
	err := ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	//t.Log(properties)
	propertiesAux, err := GetProperties("config-manager-test")
	t.Logf("%s\n", propertiesAux)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("%s\n", propertiesAux["env_name"])
	//p, err := FindProperty("Prop1")
	val, err := properties.GetValueAsString(propertiesAux, "env_name")
	if err != nil {
		t.Error(err.Error())
	}
	if val != "itdove" {
		t.Error("Expected value Val1 and get" + val)
	}
}

func TestFindProperty(t *testing.T) {
	t.Log("Entering... TestFindProperty")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	state.SetExtensionPath("../../test/resource/extensions")
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	err := ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	p, err := FindProperty("config-manager-test", "env_name")
	if err != nil {
		t.Error(err.Error())
	}
	if p == nil {
		t.Error("Can not retreive properties")
	}
	if val, ok := p.(string); ok {
		if val != "itdove" {
			t.Error("Expected value Val1 and get" + val)
		}
	} else {
		t.Error("Not a string")
	}
	p, err = FindProperty("config-manager-test", "Prop3")
	if err == nil {
		t.Error("Expected not found and found")
	}
}

func TestRemoveProperty(t *testing.T) {
	t.Log("Entering... TestRemoveProperty")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	state.SetExtensionPath("../../test/resource/extensions")
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	err := ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	err = RemoveProperty("config-manager-test", "Prop1")
	if err != nil {
		t.Error(err.Error())
	}
	p, err := FindProperty("config-manager-test", "Prop1")
	if p != nil {
		t.Error("Expected not found and found")
	}
}
