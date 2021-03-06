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
package state

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/IBM/commands-runner/api/commandsRunner/global"

	log "github.com/sirupsen/logrus"
)

type TraversePropertiesCallBack func(property map[string]interface{}, first bool, mandatory bool, parentProperty map[string]interface{}, path string, input interface{}) (err error)

func GenerateUIMetaDataTemplate(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... GenerateUIMetaDataTemplate")
	log.Debugf("extensionName=%s", extensionName)
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	if uiMetadataName == "" {
		uiMetadataName = global.DefaultUIMetaDataName
	}
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	raw, e := getUIMetadataTemplate(extensionName, uiMetadataName, langs)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadataTemplate(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataTemplate")
	cfg, err := getUIMetadataParseConfig(extensionName, uiMetadataName, langs)
	if err != nil {
		return nil, err
	}
	var path string
	outTemplate := bytes.NewBufferString("")
	groups, err := cfg.List("groups")
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			return outTemplate.Bytes(), errors.New("Expect a map[string]interface{} under groups")
		}
		if properties, ok := groupMap["properties"]; ok {
			propertiesList, ok := properties.([]interface{})
			if !ok {
				return outTemplate.Bytes(), errors.New("Expect a []interface{} under properties")
			}
			mandatory := true
			if val, ok := groupMap["mandatory"]; ok {
				mandatory = val.(bool)
			}
			err = traverseProperties(propertiesList, true, mandatory, nil, printPropertyCallBack(), path, outTemplate)
			if err != nil {
				return outTemplate.Bytes(), err
			}
		}
	}
	indentOutTemplate := bytes.NewBufferString(global.ConfigRootKey + ":\n")
	indentOutTemplate.WriteString(leftPad("configuration_name: "+uiMetadataName, 2, " ") + "\n")
	scanner := bufio.NewScanner(strings.NewReader(outTemplate.String()))
	for scanner.Scan() {
		indentOutTemplate.WriteString(leftPad(scanner.Text(), 2, " ") + "\n")
	}
	return indentOutTemplate.Bytes(), nil
}

func traverseProperties(properties []interface{}, first bool, mandatory bool, parentProperty map[string]interface{}, traversePropertiesCallBack TraversePropertiesCallBack, path string, input interface{}) error {
	for _, property := range properties {
		log.Debugf("property=%v", property)
		p, ok := property.(map[string]interface{})
		if !ok {
			return errors.New("Expect a map[string]interface{} at path " + path)
		}
		currentPropertyName, ok := p["name"]
		if !ok {
			return errors.New("Property name missing at path " + path)
		}
		log.Debug("path=" + path)
		newMandatory := mandatory
		if val, ok := p["mandatory"]; ok {
			if !val.(bool) && mandatory {
				newMandatory = false
			}
		}
		err := traversePropertiesCallBack(p, first, newMandatory, parentProperty, path, input)
		first = false
		if err != nil {
			return err
		}
		if val, ok := p["properties"]; ok {
			log.Debugf("List of properties found %v", val)
			newProperties, ok := val.([]interface{})
			if !ok {
				return errors.New("Expect an []interface{} at path: " + path)
			}
			first := false
			newPath := path
			if newPath == "" {
				newPath = currentPropertyName.(string)
			} else {
				newPath = path + "." + currentPropertyName.(string)
				if val, ok := p["type"]; ok {
					if val.(string) == "array" {
						first = true
					}
				}
			}
			err := traverseProperties(newProperties, first, newMandatory, p, traversePropertiesCallBack, newPath, input)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printPropertyCallBack() TraversePropertiesCallBack {
	log.Info("printPropertyCallBack")
	return TraversePropertiesCallBack(func(property map[string]interface{}, first bool, mandatory bool, parentProperty map[string]interface{}, path string, input interface{}) (err error) {
		var outTemplate *bytes.Buffer
		outTemplate = input.(*bytes.Buffer)
		pathArrayLen := 0
		if path != "" {
			pathArray := strings.Split(path, ".")
			log.Debugf("pathArray=%v", pathArray)
			pathArrayLen = len(pathArray)
		}
		log.Debugf("pathArrayLen=%d", pathArrayLen)
		offSet := pathArrayLen * 2
		//Build commentLine
		var commentLine string
		if val, ok := property["description"]; ok {
			commentLine = fmt.Sprintf("%s", leftPad("# "+val.(string), offSet, " "))
		} else {
			commentLine = fmt.Sprintf("%s", leftPad("# No description", offSet, " "))
		}
		//Search sample value
		var sampleValue interface{}
		if val, ok := property["sample_value"]; ok {
			sampleValue = val
		} else {
			if val, ok := property["items"]; ok {
				options := val.([]interface{})
				sampleValue = "Possible values are: "
				for index := range options {
					option := options[index].(map[string]interface{})
					sampleValue = sampleValue.(string) + option["label"].(string)
					if index < len(options)-1 {
						sampleValue = sampleValue.(string) + ", "
					}
				}
			}
		}
		//Add sampleValue to comment
		if sampleValue != nil {
			switch sampleValue.(type) {
			case string:
				commentLine = commentLine + fmt.Sprintf(" - sample_value: %s", sampleValue)
			case int:
				commentLine = commentLine + fmt.Sprintf(" - sample_value: %d", sampleValue)
			case bool:
				commentLine = commentLine + fmt.Sprintf(" - sample_value: %t", sampleValue)
			default:
			}
		}
		// write comment line of not empty
		if commentLine != "" {
			commentLine = strings.Replace(commentLine, "\n", "\n# ", -1) + fmt.Sprint("\n")
			writeBuffer(outTemplate, mandatory, commentLine)
		}
		//Search default value
		var defaultValue interface{}
		if val, ok := property["default"]; ok {
			defaultValue = val
		}
		nameLine := ""
		if _, ok := property["properties"]; ok {
			nameLine = fmt.Sprintf("%s:\n", leftPad(property["name"].(string), offSet, " "))
		} else {
			//Create line with default
			switch defaultValue.(type) {
			case string:
				if !mandatory {
					defaultValue = strings.Replace(defaultValue.(string), "\n", "\n# ", -1)
				}
				nameLine = fmt.Sprintf("%s: \"%s\"\n", leftPad(property["name"].(string), offSet, " "), defaultValue)
			case int:
				nameLine = fmt.Sprintf("%s: %d\n", leftPad(property["name"].(string), offSet, " "), defaultValue)
			case bool:
				nameLine = fmt.Sprintf("%s: %t\n", leftPad(property["name"].(string), offSet, " "), defaultValue)
			default:
				//if no default found then create line with sample value as comment
				switch sampleValue.(type) {
				case string:
					if !mandatory {
						sampleValue = strings.Replace(sampleValue.(string), "\n", "\n# ", -1)
					}
					nameLine = fmt.Sprintf("%s: # %s\n", leftPad(property["name"].(string), offSet, " "), sampleValue)
				case int:
					nameLine = fmt.Sprintf("%s: # %d\n", leftPad(property["name"].(string), offSet, " "), sampleValue)
				case bool:
					nameLine = fmt.Sprintf("%s: # %t\n", leftPad(property["name"].(string), offSet, " "), sampleValue)
				default:
					nameLine = fmt.Sprintf("%s: # %s\n", leftPad(property["name"].(string), offSet, " "), "No default/sample_value provided or unknown default/sample_value type")
				}
			}
			log.Debug("nameLine1:" + nameLine)
			if val, ok := parentProperty["type"]; ok {
				propertyType := val.(string)
				log.Debugf("propertyType=%s", propertyType)
				log.Debugf("first=%t", first)
				if propertyType == "array" && first {
					nameLine = strings.Replace(nameLine, "  "+property["name"].(string), "- "+property["name"].(string), 1)
					log.Debug("nameLine2:" + nameLine)
				}
			}
		}
		writeBuffer(outTemplate, mandatory, nameLine)
		log.Info(outTemplate.String())
		return nil
	})
}

func leftPad(s string, nb int, char string) string {
	b := bytes.NewBufferString("")
	for i := 0; i < nb; i++ {
		b.WriteString(char)
	}
	b.WriteString(s)
	return b.String()
}

func writeBuffer(buffer *bytes.Buffer, mandatory bool, line string) (n int, err error) {
	if !mandatory {
		line = "# " + line
	}
	return buffer.WriteString(line)
}
