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
/* File contains generic code to manage the uiconfig */
package config

import (
	"encoding/base64"
	"errors"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/properties"
	"github.com/IBM/commands-runner/api/commandsRunner/state"

	"github.com/olebedev/config"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Config struct {
	Properties properties.Properties `json:"config" yaml:"config"`
}

var props properties.Properties

//Initialize the properties map
func init() {
	props = make(properties.Properties)
}

/* Search a property in a ui config object, return an error if not found
 */
func SearchUIConfigProperty(extensionName, uiMetaDataName string, name string) (*config.Config, error) {
	log.Debug("Entering... searchUIConfigProperty:" + name)
	b, err := state.GetUIMetaDataConfig(extensionName, uiMetaDataName, []string{global.DefaultLanguage})
	if err != nil {
		return nil, err
	}
	uiConfigDeploy, err := config.ParseJson(string(b))
	//Search list of groups
	groups, err := uiConfigDeploy.List("groups")
	log.Debug("Group size:" + strconv.Itoa(len(groups)))
	if err != nil {
		log.Debug("Search groups error:" + err.Error())
		return nil, err
	}
	//Loop on all groups
	for groupIndex, _ := range groups {
		//Retrieve the properties attribute from the group
		properties, errProps := uiConfigDeploy.List("groups." + strconv.Itoa(groupIndex) + ".properties")
		log.Debug("properties size:" + strconv.Itoa(len(properties)))
		if errProps != nil {
			log.Debug("Search properties error:" + errProps.Error())
			return nil, errProps
		}
		//Loop on all properties
		for propertiesIndex, _ := range properties {
			//Retrieve the property
			property, errProp := uiConfigDeploy.Get("groups." + strconv.Itoa(groupIndex) + ".properties." + strconv.Itoa(propertiesIndex))
			if errProp != nil {
				log.Debug("Search property error:" + errProp.Error())
				return nil, errProp
			}
			//Retrieve the property "nane"
			nameFound, errName := property.String("name")
			log.Debug("UI Property found:" + nameFound)
			if errName != nil {
				log.Debug("Search nameFound error:" + errName.Error())
				return nil, errName
			}
			//Check if it is the searched name
			if nameFound == name {
				log.Debug("Found Property:" + name)
				return property, nil
			}
		}
	}
	return nil, errors.New(name + " not found!")
}

/*
Set the config path and read the properties
If property file not present, an empty file is created
The Listener stops if the file can not be created
*/
func SetConfigPath(configDirectoryP string) {
	//Build the uiconfig path
	global.ConfigDirectory = configDirectoryP
}

/*
Set the config file name
*/
func SetConfigFileName(configFileName string) {
	global.ConfigYamlFileName = configFileName
}

/*
Set the config root key
*/
func SetConfigRootKey(rootKey string) {
	global.ConfigRootKey = rootKey
}

/*
Set the client path
*/
// func SetClientPath(clientPath string) {
// 	global.ClientPath = clientPath
// }

/*
Search the configuration_name property
*/
func GetConfigurationName(extensionName string) (string, error) {
	log.Debug("Entering... GetConfigurationName")
	props, err := properties.ReadProperties(extensionName)
	if err != nil {
		return "", err
	}
	return properties.GetValueAsString(props, "configuration_name")

}

/*
Save the property map in the property file
Reread the file afterward
*/
func SetProperties(extensionName string, ps properties.Properties) error {
	log.Debug("Entering... SetProperties")
	registered := state.IsExtensionRegistered(extensionName)
	if !registered {
		err := errors.New("Extension " + extensionName + "not registered yet")
		log.Debug(err.Error())
		return err
	}
	err := properties.WriteProperties(extensionName, ps)
	if err != nil {
		return err
	}
	props, err = properties.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	return nil
}

/*
Encode decode properties
*/
func PropertiesEncodeDecode(extensionName string, uiMetadataName string, ps properties.Properties, encode bool) (properties.Properties, error) {
	log.Debug("Entering in... PropertiesEncodeDecode")
	pss := make(properties.Properties)
	for key, val := range ps {
		uiProperty, err := SearchUIConfigProperty(extensionName, uiMetadataName, key)
		if err == nil {
			s, _ := config.RenderYaml(uiProperty)
			log.Debug("uiProperty:" + s)
			uiPropertyEncode, errEncode := uiProperty.String("encode")
			log.Debug("Encode:" + uiPropertyEncode)
			if errEncode == nil {
				log.Debug("uiPropertyEncode:" + uiPropertyEncode)
				val, err = PropertyEncodeDecode(uiPropertyEncode, val.(string), encode)
				if err != nil {
					return nil, err
				}
			}
		}
		pss[key] = val
	}
	return pss, nil
}

func PropertyEncodeDecode(encodingType string, val string, encode bool) (string, error) {
	log.Debug("Entering in... propertyEncodeDecode")
	//Decode if needed, data string contains the value
	var result string
	switch encodingType {
	case "base64":
		if encode {
			result = base64.StdEncoding.EncodeToString([]byte(val))
		} else {
			dataDecode, errDecode := base64.StdEncoding.DecodeString(val)
			if errDecode != nil {
				return "", errDecode
			}
			result = string(dataDecode)
		}
	default:
		result = val
	}
	log.Debug("Result:" + result)
	return result, nil
}

/*
Read the property file and populate the map.
If read can not be done, the error is forwarded
*/
func GetProperties(extensionName string) (properties.Properties, error) {
	log.Debug("Entering in... GetProperties")
	log.Debug("extensionName:" + extensionName)
	properties, e := properties.ReadProperties(extensionName)
	if e != nil {
		return nil, e
	}
	return properties, nil
}

/*
Remove a property from the map
*/
func RemoveProperty(extensionName string, key string) error {
	propertiesAux, err := properties.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	delete(propertiesAux, key)
	SetProperties(extensionName, propertiesAux)
	propertiesAux, err = properties.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	return nil
}

/*
Search for a given property
*/
func FindProperty(extensionName string, key string) (properties.Properties, error) {
	var pss properties.Properties
	pss = make(properties.Properties)
	properties, err := properties.ReadProperties(extensionName)
	if err != nil {
		return nil, err
	}
	if p, ok := properties[key]; ok {
		pss["name"] = key
		pss["value"] = p
		return pss, nil
	}
	err = errors.New("Property " + key + " not found")
	return nil, err
}

/*
Add a property
*/
func AddProperty(extensionName string, key string, value interface{}) error {
	var err error
	props, err = properties.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	props[key] = value
	SetProperties(extensionName, props)
	return nil
}
