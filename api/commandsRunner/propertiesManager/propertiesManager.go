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

package propertiesManager

import (
	"errors"
	"io/ioutil"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

type Properties map[string]interface{}

//get the cf-pie path
func GetConfigPath(extensionName string) string {
	var configPath string
	if extensionName == global.CommandsRunnerStatesName || extensionName == "" {
		configPath = extensionManager.GetRootExtensionPath(global.ConfigDirectory, extensionName)
	} else {
		configPath = extensionManager.GetRootExtensionPath(extensionManager.GetExtensionPath(), extensionName)
	}
	return configPath
}

func logProperties(ps Properties) {
	for key, _ := range ps {
		if !strings.Contains(key, "password") &&
			!strings.Contains(key, "key") &&
			!strings.Contains(key, "cert") {
			val, err := GetValueAsString(ps, key)
			if err != nil {
				log.Debug(key + " not a string")
			} else {
				log.Debugf("%s: %s", key, val)
			}

		} else {
			log.Debugf("%s: *****", key)
		}
	}
}

/*
Read the property file and populate the properties map
If the file is not present or can not be read an error is raised
*/
func ReadProperties(extensionName string) (Properties, error) {
	log.Debug("Entering... readProperties")
	dataDirectory := GetConfigPath(extensionName)
	//properties = make(global.Properties)
	log.Debugf("dataDirectory:%s\n", dataDirectory)
	raw, e := ioutil.ReadFile(dataDirectory + global.UIConfigYamlFileName)
	//log.Debugf("\n%s", string(raw))
	if e != nil {
		return nil, errors.New("Unable to read " + dataDirectory + global.UIConfigYamlFileName + " " + e.Error())
	}
	//	var bmxConfig BMXConfig
	uiConfigCfg, err := config.ParseYamlBytes(raw)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	var properties Properties
	properties, err = uiConfigCfg.Map("uiconfig")
	//	err := yaml.Unmarshal(raw, &bmxConfig)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	//	log.Debug("data:\n" + string(raw))
	logProperties(properties)
	return properties, err
}

func RenderProperties(ps Properties) (string, error) {
	log.Debug("Entering... renderProperties")
	uiConfigYaml, err := config.ParseYaml("uiconfig:")
	if err != nil {
		log.Debug("parse:" + err.Error())
	}
	err = uiConfigYaml.Set("uiconfig", ps)
	if err != nil {
		log.Debug("Set:" + err.Error())
	}
	out, err := config.RenderYaml(uiConfigYaml.Root)
	if err != nil {
		log.Debug("Set:" + err.Error())
		return "", err
	}
	return string(out), nil
}

func WriteProperties(extensionName string, ps Properties) error {
	log.Debug("Entering... writeProperties")
	dataDirectory := GetConfigPath(extensionName)
	log.Debug("dataDirectory:" + dataDirectory)
	propertiesYaml, err := RenderProperties(ps)
	//	log.Debug("propertiesYaml:\n" + propertiesYaml)
	err = ioutil.WriteFile(dataDirectory+global.UIConfigYamlFileName, []byte(propertiesYaml), 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetValueAsString(ps Properties, key string) (string, error) {
	if val, ok := ps[key]; ok {
		val, err := ConvertToString(val)
		if err != nil {
			return "", err
		}
		return val, nil
	}
	return "", nil
}

func ConvertToString(data interface{}) (string, error) {
	if data == nil {
		return "", errors.New("data nil")
	}
	if val, ok := data.(string); ok {
		return val, nil
	} else {
		return "", errors.New("Not a string:" + reflect.TypeOf(data).String())
	}
}

func GetValueAsBool(ps Properties, key string) (bool, error) {
	if val, ok := ps[key]; ok {
		val, err := ConvertToBool(val)
		if err != nil {
			return false, err
		}
		return val, nil
	}
	return false, nil
}

func ConvertToBool(data interface{}) (bool, error) {
	if val, ok := data.(bool); ok {
		return val, nil
	} else {
		return false, errors.New("Not a boolean")
	}
}

func AddError(ps Properties, key string, msgType string, msg string) Properties {
	var property Properties
	//Do nothing if it is already a map
	if val, ok := ps[key].(Properties); ok {
		log.Debug("Already a properties...")
		return val
	}
	property = make(Properties)
	property["value"] = ps[key]
	property["message_type"] = msgType
	property["message"] = msg
	return property
}
