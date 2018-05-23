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
/* File contains generic code to manage the uiconfig */
package configManager

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/propertiesManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/uiConfigManager"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"

	"github.com/olebedev/config"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Config struct {
	Properties propertiesManager.Properties `json:"uiconfig" yaml:"uiconfig"`
}

var properties propertiesManager.Properties

//var uiCFDeploy UIConfig
//var uiCFDeploy *config.Config

//Initialize the properties map
func init() {
	properties = make(propertiesManager.Properties)
}

/* Search a property in a ui config object, return an error if not found
 */
func searchUIConfigProperty(extensionName, name string) (*config.Config, error) {
	log.Debug("Entering... searchProperty:" + name)
	b, err := uiConfigManager.GetUIConfig(extensionName)
	if err != nil {
		return nil, err
	}
	uiCFDeploy, err := config.ParseJson(string(b))
	//Search list of groups
	groups, err := uiCFDeploy.List("groups")
	log.Debug("Group size:" + strconv.Itoa(len(groups)))
	if err != nil {
		log.Debug("Search groups error:" + err.Error())
		return nil, err
	}
	//Loop on all groups
	for groupIndex, _ := range groups {
		//Retrieve the properties attribute from the group
		properties, errProps := uiCFDeploy.List("groups." + strconv.Itoa(groupIndex) + ".properties")
		log.Debug("properties size:" + strconv.Itoa(len(properties)))
		if errProps != nil {
			log.Debug("Search properties error:" + errProps.Error())
			return nil, errProps
		}
		//Loop on all properties
		for propertiesIndex, _ := range properties {
			//Retrieve the property
			property, errProp := uiCFDeploy.Get("groups." + strconv.Itoa(groupIndex) + ".properties." + strconv.Itoa(propertiesIndex))
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
	uiConfigPathYaml := configDirectoryP + string(filepath.Separator) + global.UIConfigYamlFileName
	log.Debugf("uiConfigPathYaml:%s", uiConfigPathYaml)
	//Read the current properties
	propertiesManager.ReadProperties(global.CommandsRunnerStatesName)
}

/*
Save the property map in the property file
Reread the file afterward
*/
func SetProperties(extensionName string, ps propertiesManager.Properties) error {
	log.Debug("Entering... SetProperties")
	registered := extensionManager.IsExtensionRegistered(extensionName)
	if !registered {
		err := errors.New("Extension " + extensionName + "not registered yet")
		log.Debug(err.Error())
		return err
	}
	err := propertiesManager.WriteProperties(extensionName, ps)
	if err != nil {
		return err
	}
	properties, err = propertiesManager.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	return nil
}

func PropertiesEncodeDecode(extensionName string, ps propertiesManager.Properties, encode bool) (propertiesManager.Properties, error) {
	log.Debug("Entering in... PropertiesEncodeDecode")
	pss := make(propertiesManager.Properties)
	for key, val := range ps {
		uiProperty, err := searchUIConfigProperty(extensionName, key)
		if err == nil {
			s, _ := config.RenderYaml(uiProperty)
			log.Debug("uiProperty:" + s)
			uiPropertyEncode, errEncode := uiProperty.String("encode")
			log.Debug("Encode:" + uiPropertyEncode)
			if errEncode == nil {
				log.Debug("uiPropertyEncode:" + uiPropertyEncode)
				val, err = propertyEncodeDecode(uiPropertyEncode, val.(string), encode)
				if err != nil {
					return nil, err
				}
			}
		}
		pss[key] = val
	}
	return pss, nil
}

func PropertiesEncodeDecodeValidate(extensionName string, ps propertiesManager.Properties, encode bool) (*propertiesManager.Properties, error) {
	log.Debug("Entering in... PropertiesEncodeDecode")
	pss := make(propertiesManager.Properties)
	for key, val := range ps {
		uiProperty, err := searchUIConfigProperty(extensionName, key)
		if err == nil {
			s, _ := config.RenderYaml(uiProperty)
			log.Debug("uiProperty:" + s)
			uiPropertyEncode, errEncode := uiProperty.String("encode")
			log.Debug("Encode:" + uiPropertyEncode)
			if errEncode == nil {
				log.Debug("uiPropertyEncode:" + uiPropertyEncode)
				valToConvert := val.(propertiesManager.Properties)["value"].(string)
				converted, err := propertyEncodeDecode(uiPropertyEncode, valToConvert, encode)
				if err != nil {
					return nil, err
				}
				val.(propertiesManager.Properties)["value"] = converted
			}
		}
		pss[key] = val
	}
	return &pss, nil
}

func propertyEncodeDecode(encodingType string, val string, encode bool) (string, error) {
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
func GetProperties(extensionName string) (propertiesManager.Properties, error) {
	log.Debug("Entering in... GetProperties")
	log.Debug("extensionName:" + extensionName)
	properties, e := propertiesManager.ReadProperties(extensionName)
	if e != nil {
		return nil, e
	}
	return properties, nil
}

/*
Remove a property from the map
*/
func RemoveProperty(extensionName string, key string) error {
	properties, err := propertiesManager.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	delete(properties, key)
	SetProperties(extensionName, properties)
	properties, err = propertiesManager.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	return nil
}

/*
Search for a given property
*/
func FindProperty(extensionName string, key string) (interface{}, error) {
	properties, err := propertiesManager.ReadProperties(extensionName)
	if err != nil {
		return nil, err
	}
	if p, ok := properties[key]; ok {
		return p, nil
	}
	err = errors.New("Property " + key + " not found")
	return nil, err
}

func AddProperty(extensionName string, key string, value interface{}) error {
	var err error
	properties, err = propertiesManager.ReadProperties(extensionName)
	if err != nil {
		return err
	}
	properties[key] = value
	SetProperties(extensionName, properties)
	return nil
}
