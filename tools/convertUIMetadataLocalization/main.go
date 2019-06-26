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
package main

import (
	"errors"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
	cli "gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	//	log.SetLevel(log.DebugLevel)
	var sourceFile string
	var destFile string
	var translationFile string

	convertUIMetadataForLocationzation := func(c *cli.Context) error {
		err := convertUIMetadataForLocationzation(sourceFile, destFile, translationFile)
		if err != nil {
			log.Error(err.Error())
		}
		return err
	}

	app := cli.NewApp()
	//Overwrite some app parameters
	app.Usage = "client ..."
	app.Version = "1.0.0"
	app.Description = "Sample client"

	//Enrich with extra client commands
	app.Commands = []cli.Command{
		{
			Name:   "convert",
			Action: convertUIMetadataForLocationzation,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "sourceFile, s",
					Usage:       "The path to the uimetadata file to convert",
					Destination: &sourceFile,
				},
				cli.StringFlag{
					Name:        "destFile, d",
					Usage:       "The path to the generated destination file",
					Destination: &destFile,
				},
				cli.StringFlag{
					Name:        "translation, t",
					Usage:       "The path to the generated translation file",
					Destination: &translationFile,
				},
			},
		},
	}

	//Run the command
	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}
}

func convertUIMetadataForLocationzation(sourceFile string, destFile string, translationFile string) error {
	var translationsMap map[string]string
	translationsMap = make(map[string]string, 0)
	log.Debug("Read translation File:" + translationFile)
	if _, err := os.Stat(translationFile); err == nil {
		log.Debug("Translation File exists")
		translationData, err := ioutil.ReadFile(translationFile)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(translationData, &translationsMap)
		if err != nil {
			return err
		}
		log.Debug("Translation File contains:" + string(translationData))
	} else {
		log.Debug("Translation File doesn't exist")
	}

	log.Debug("Read file:" + sourceFile)
	manifest, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	log.Debug("Parse:" + string(manifest))
	cfgSource, err := config.ParseYaml(string(manifest))
	if err != nil {
		return err
	}
	log.Debug("Get ui_metadata")
	cfg, err := cfgSource.Get("ui_metadata")
	if err != nil {
		return err
	}
	err = translateUIMetadata(cfg, "en", translationsMap)
	if err != nil {
		return err
	}
	data, err := config.RenderYaml(cfgSource.Root)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	w, err := os.Create(translationFile)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(w)
	err = encoder.Encode(translationsMap)
	if err != nil {
		return err
	}
	w.Close()
	// err = ioutil.WriteFile(translationFile, outTemplate.Bytes(), 0644)
	return ioutil.WriteFile(destFile, []byte(data), 0644)
}

func translateUIMetadata(cfg *config.Config, lang string, translationsMap map[string]string) error {
	uiMetadataNames, err := cfg.Map("")
	if err != nil {
		return err
	}
	for uiMetadataKey, val := range uiMetadataNames {
		uiMetadataNameMap := val.(map[string]interface{})
		if uiMetadataNameLabel, ok := uiMetadataNameMap["label"]; ok {
			uiMetadataNameMap["label"] = uiMetadataKey + ".label"
			addMessage(translationsMap, uiMetadataNameMap, "label", uiMetadataKey+".label", uiMetadataNameLabel)
		}
		groups := uiMetadataNameMap["groups"].([]interface{})
		for _, group := range groups {
			groupMap, ok := group.(map[string]interface{})
			if !ok {
				return errors.New("Expect a map[string]interface{} under groups")
			}
			if groupLabel, ok := groupMap["label"]; ok {
				groupMap["label"] = uiMetadataKey + "." + groupMap["name"].(string) + ".label"
				addMessage(translationsMap, groupMap, "label", uiMetadataKey+"."+groupMap["name"].(string)+".label", groupLabel)
			}
			if properties, ok := groupMap["properties"]; ok {
				propertiesList, ok := properties.([]interface{})
				if !ok {
					return errors.New("Expect a []interface{} under properties")
				}
				err = translateProperties(propertiesList, uiMetadataKey+"."+groupMap["name"].(string), lang, translationsMap)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func translateProperties(properties []interface{}, path string, lang string, translationsMap map[string]string) error {
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
		translateProperty(p, path+"."+currentPropertyName.(string), lang, translationsMap)
		if val, ok := p["properties"]; ok {
			log.Debugf("List of properties found %v", val)
			newProperties, ok := val.([]interface{})
			if !ok {
				return errors.New("Expect an []interface{} at path: " + path)
			}
			newPath := path

			newPath = path + "." + currentPropertyName.(string)

			err := translateProperties(newProperties, newPath, lang, translationsMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func translateProperty(property map[string]interface{}, path string, lang string, translationsMap map[string]string) {
	var propertyPath string
	if val, ok := property["description"]; ok {
		propertyPath = path + ".description"
		addMessage(translationsMap, property, "description", propertyPath, val)
	}
	if val, ok := property["label"]; ok {
		propertyPath = path + ".label"
		addMessage(translationsMap, property, "label", propertyPath, val)
	}
	if val, ok := property["validation_error_message"]; ok {
		propertyPath = path + ".validation_error_message"
		addMessage(translationsMap, property, "validation_error_message", propertyPath, val)
	}
	if val, ok := property["items"]; ok {
		options := val.([]interface{})
		for index := range options {
			option := options[index].(map[string]interface{})
			propertyPath = path + ".items." + option["id"].(string)
			addMessage(translationsMap, option, "label", propertyPath, option["label"])
		}
	}
	if val, ok := property["sample_value"]; ok {
		switch val.(type) {
		case string:
			propertyPath = path + ".sample_value"
			addMessage(translationsMap, property, "sample_value", propertyPath, val)
		default:
		}
	}

}

func addMessage(translationsMap map[string]string, property map[string]interface{}, key string, propertyPath string, value interface{}) error {
	if _, ok := translationsMap[propertyPath]; !ok {
		property[key] = propertyPath
		translationsMap[propertyPath] = value.(string)
		log.Info("Destination and translation files updated with: " + propertyPath + ":" + value.(string))
		return nil
	}
	log.Info("Destination and translation files not updated because it already contains key: " + propertyPath + " already exists")
	return nil
}
