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
package configManager

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/propertiesManager"
	"github.ibm.com/IBMPrivateCloud/cfp-config-manager/api/commandsRunnerCustom/globalCustom"
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

var bmxConfigString string = "uiconfig:\n  env_name: \"itdove\"\n  host_directory: \"/itdove/data\"\n  subnet: \"192.168.100.0/24\""

//var global.ConfigDirectory string = "../../test/resource"
//var properties propertiesManager.Properties

func TestSetConfigPathNotDone(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering... TestSetConfigPathNotDone")
	_, err := GetProperties(global.CommandsRunnerStatesName)
	if err == nil {
		t.Error("An error should be raised as the SetConfigPath is not yet set")
	}
}

func TestSetConfigPath(t *testing.T) {
	t.Log("Entering... TestSetConfigPath")
	global.ConfigDirectory = "../../test/resource"
	os.Remove(global.ConfigDirectory)
	SetConfigPath(global.ConfigDirectory)
}

func TestSetProperties(t *testing.T) {
	t.Log("Entering... TestSetpropertiesManager.Properties")
	properties = make(propertiesManager.Properties)
	global.ConfigDirectory = "../../test/resource"
	SetConfigPath(global.ConfigDirectory)
	//	t.Error(global.ConfigDirectory)
	os.MkdirAll(global.ConfigDirectory, 0744)
	properties["Prop3"] = "Val3"
	properties["Prop4"] = "Val4"
	properties["subnet"] = "192.168.100.0/24"
	err := SetProperties(global.CommandsRunnerStatesName, properties)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetProperties(t *testing.T) {
	t.Log("Entering... TestGetpropertiesManager.Properties")
	t.Logf("%s\n", bmxConfigString)
	global.ConfigDirectory = "../../test/resource"
	dataDirectory := extensionManager.GetRootExtensionPath(global.ConfigDirectory, global.CommandsRunnerStatesName)
	t.Log("dataDirectory:" + dataDirectory)
	err := ioutil.WriteFile(dataDirectory+global.UIConfigYamlFileName, []byte(bmxConfigString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	//t.Log(properties)
	properties, err := GetProperties(global.CommandsRunnerStatesName)
	t.Logf("%s\n", properties)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("%s\n", properties["env_name"])
	//p, err := FindProperty("Prop1")
	val, err := propertiesManager.GetValueAsString(properties, "env_name")
	if err != nil {
		t.Error(err.Error())
	}
	if val != "itdove" {
		t.Error("Expected value Val1 and get" + val)
	}
}

func TestFindProperty(t *testing.T) {
	t.Log("Entering... TestFindProperty")
	t.Logf("%s\n", bmxConfigString)
	global.ConfigDirectory = "../../test/resource"
	dataDirectory := extensionManager.GetRootExtensionPath(global.ConfigDirectory, global.CommandsRunnerStatesName)
	err := ioutil.WriteFile(dataDirectory+global.UIConfigYamlFileName, []byte(bmxConfigString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	p, err := FindProperty(global.CommandsRunnerStatesName, "env_name")
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
	p, err = FindProperty(global.CommandsRunnerStatesName, "Prop3")
	if err == nil {
		t.Error("Expected not found and found")
	}
}

func TestRemoveProperty(t *testing.T) {
	t.Log("Entering... TestRemoveProperty")
	t.Logf("%s\n", bmxConfigString)
	global.ConfigDirectory = "../../test/resource"
	dataDirectory := extensionManager.GetRootExtensionPath(global.ConfigDirectory, global.CommandsRunnerStatesName)
	err := ioutil.WriteFile(dataDirectory+globalCustom.UIConfigJsonFileName, []byte(bmxConfigString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	err = RemoveProperty(global.CommandsRunnerStatesName, "Prop1")
	if err != nil {
		t.Error(err.Error())
	}
	p, err := FindProperty(global.CommandsRunnerStatesName, "Prop1")
	if p != nil {
		t.Error("Expected not found and found")
	}
}
