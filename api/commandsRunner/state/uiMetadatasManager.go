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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/i18n/i18nUtils"

	"github.com/olebedev/config"
)

func GetUIMetaDataConfigs(extensionName string, namesOnly bool, langs []string) ([]byte, error) {
	log.Debug("Entering in... GetUIMetaDataConfig")
	log.Debugf("extensionName=%s", extensionName)
	var raw []byte
	var e error
	if extensionName == "" {
		raw, e = getUIMetadataAllConfigs(namesOnly, langs)
	} else {
		raw, e = getUIMetadataConfigs(extensionName, namesOnly, langs)
	}
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadataAllConfigs(namesOnly bool, langs []string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataAllConfigs")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	cfg, err := getUIMetadataAllConfigsCFg(namesOnly, langs)
	if err == nil {
		errUIConfigFileCfg = config.Set(uiConfigFileCfg.Root, "ui_metadata", cfg.Root)
		if errUIConfigFileCfg != nil {
			return nil, errUIConfigFileCfg
		}
		out, err := config.RenderJson(uiConfigFileCfg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, err
}

func getUIMetadataConfigs(extensionName string, namesOnly bool, langs []string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataConfigs")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	cfg, err := getUIMetadataConfigsCfg(extensionName, namesOnly, langs)
	if err == nil && cfg != nil {
		err = uiConfigFileCfg.Set("ui_metadata", cfg.Root)
		if err != nil {
			return nil, err
		}
		out, err := config.RenderJson(uiConfigFileCfg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	if cfg == nil {
		err = errors.New("No ui configuration available for " + extensionName)
	}
	return nil, err
}

func getUIMetadataAllConfigsCFg(namesOnly bool, langs []string) (*config.Config, error) {
	log.Debug("Entering in... getUIMetadataAllConfigsCFg")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("dummy:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	extensions, err := ListExtensions("", false)
	if err != nil {
		return nil, err
	}
	for key := range extensions.Extensions {
		log.Debug("extension-name:" + key)
		cfg, err := getUIMetadataConfigsCfg(key, namesOnly, langs)
		log.Debug(cfg)
		if err != nil {
			if cfg == nil {
				continue
			} else {
				return nil, err
			}
		}
		log.Debug(cfg)
		log.Debug(key)
		errExtensionCfg := config.Set(uiConfigFileCfg.Root, "ui_metadata."+key, cfg.Root)
		if errExtensionCfg != nil {
			log.Debug("----------- ERROR-------" + err.Error())
			return nil, err
		}
		log.Debug(uiConfigFileCfg)
	}
	log.Debug("uiConfigFileCfg")
	log.Debug(uiConfigFileCfg)

	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	return uiConfigFileCfg.Get("ui_metadata")
}

func getUIMetadataConfigsCfg(extensionName string, namesOnly bool, langs []string) (*config.Config, error) {
	log.Debug("Entering in... getUIMetadataConfigsCfg")
	cfg, err := getUIMetadataParseConfigs(extensionName, langs)
	log.Debug("")
	if err == nil {
		log.Debug("No ERROR")
		log.Debug("CFG not empty")
		if namesOnly {
			cfg, err = getUIMetadataNameOnly(cfg)
			if err != nil {
				return nil, err
			}
			log.Debug("cfg")
			log.Debug(cfg)
		}
	}
	return cfg, err
}

func getUIMetadataParseConfigs(extensionName string, langs []string) (cfg *config.Config, err error) {
	log.Debug("Entering in... getUIMetadataParseConfigs")
	rootPath := GetExtensionsPathCustom()
	embeddedExtension, _ := IsEmbeddedExtension(extensionName)
	if embeddedExtension {
		rootPath = GetExtensionsPathEmbedded()
	}

	cfg, _, err = getUIMetadataTranslatedAndWrite(filepath.Join(rootPath, extensionName), langs)
	return cfg, err
}

func getUIMetadataTranslatedAndWrite(extensionPath string, langs []string) (cfg *config.Config, messagesNotFound []string, err error) {
	log.Debug("Entering in... getUIMetadataTranslatedAndWrite")
	cfg, _, err = GetUIMetadataTranslated(extensionPath, langs)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		return cfg, nil, err
	}
	log.Debug("Set ui_metadata")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return cfg, nil, errUIConfigFileCfg
	}
	errUIConfigFileCfg = config.Set(uiConfigFileCfg.Root, "ui_metadata", cfg.Root)
	if errUIConfigFileCfg != nil {
		return cfg, nil, errUIConfigFileCfg
	}
	out, err := config.RenderYaml(uiConfigFileCfg.Root)
	if err != nil {
		return cfg, nil, err
	}
	translatedFileName := filepath.Join(extensionPath, "ui_metadata."+langs[0]+".yml")
	log.Debug("Translate into file:" + translatedFileName)
	err = ioutil.WriteFile(translatedFileName, []byte(out), 0644)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Translated file written: " + translatedFileName)
	return cfg, nil, err
}

func GetUIMetadataTranslated(extensionPath string, langs []string) (cfg *config.Config, messagesNotFound []string, err error) {
	log.Debug("Entering in... GetUIMetadataTranslated")
	var supportedLang string
	for _, lang := range langs {
		log.Debug("Processing lang:" + lang)
		translatedFileName := filepath.Join(extensionPath, "ui_metadata."+lang+".yml")
		log.Debug("Check if file exists:" + translatedFileName)
		if _, err := os.Stat(translatedFileName); err == nil {
			log.Debug("File exists:" + translatedFileName)
			uimetadata, err := ioutil.ReadFile(translatedFileName)
			if err != nil {
				return nil, nil, err
			}
			cfg, err = config.ParseYaml(string(uimetadata))
			if err != nil {
				return nil, nil, err
			}
			cfg, err = cfg.Get("ui_metadata")
			return cfg, nil, err
		}
		if i18nUtils.IsSupportedLanguage(lang) {
			supportedLang = lang
			log.Debug(lang + " is supported but the file doesn't exist yet")
			//The language is supported but the file doesn't exist yet.
			break
		}
	}
	//File doesn't exist yet, needs to be created
	filePath := filepath.Join(extensionPath, global.DefaultExtenstionManifestFile)
	log.Debug("Read file:" + filePath)
	manifest, err := ioutil.ReadFile(filePath)
	if err != nil {
		return cfg, nil, err
	}
	log.Debug("Parse:" + string(manifest))
	cfg, err = config.ParseYaml(string(manifest))
	if err != nil {
		return cfg, nil, err
	}
	log.Debug("Get ui_metadata")
	cfg, err = cfg.Get("ui_metadata")
	if err != nil {
		return cfg, nil, err
	}
	messagesNotFound = make([]string, 0)
	messagesNotFound, err = translateUIMetadata(cfg, supportedLang)
	if err != nil {
		return cfg, messagesNotFound, err
	}
	return cfg, messagesNotFound, err
}

func translateUIMetadata(cfg *config.Config, lang string) ([]string, error) {
	log.Debug("Entering in... translateUIMetadata")
	messagesNotFound := make([]string, 0)
	var path string
	uiMetadataNames, err := cfg.Map("")
	if err != nil {
		return nil, err
	}
	for _, val := range uiMetadataNames {
		uiMetadataNameMap := val.(map[string]interface{})
		if uiMetadataNameLabel, ok := uiMetadataNameMap["label"]; ok {
			uiMetadataNameMap["label"], _ = i18nUtils.Translate(uiMetadataNameLabel.(string), uiMetadataNameLabel.(string), []string{lang})
			if uiMetadataNameMap["label"].(string) == uiMetadataNameLabel.(string) {
				log.Warning("message '" + uiMetadataNameLabel.(string) + "' not found")
				messagesNotFound = append(messagesNotFound, uiMetadataNameLabel.(string))
			}
		}
		groups := uiMetadataNameMap["groups"].([]interface{})
		for _, group := range groups {
			groupMap, ok := group.(map[string]interface{})
			if !ok {
				return messagesNotFound, errors.New("Expect a map[string]interface{} under groups")
			}
			if groupLabel, ok := groupMap["label"]; ok {
				groupMap["label"], _ = i18nUtils.Translate(groupLabel.(string), groupLabel.(string), []string{lang})
				if groupMap["label"].(string) == groupLabel.(string) {
					log.Warning("message '" + groupLabel.(string) + "' not found")
					messagesNotFound = append(messagesNotFound, groupLabel.(string))
				}
			}
			if properties, ok := groupMap["properties"]; ok {
				propertiesList, ok := properties.([]interface{})
				if !ok {
					return messagesNotFound, errors.New("Expect a []interface{} under properties")
				}
				m, err := translateProperties(propertiesList, true, nil, path, lang, messagesNotFound)
				if err != nil {
					return nil, err
				}
				messagesNotFound = append(messagesNotFound, m...)
			}
		}
	}
	return messagesNotFound, nil
}

func translateProperties(properties []interface{}, first bool, parentProperty map[string]interface{}, path string, lang string, messagesAlreadyNotFound []string) ([]string, error) {
	log.Debug("Entering in... translateProperties")
	messagesNotFound := make([]string, 0)
	for _, property := range properties {
		log.Debugf("property=%v", property)
		p, ok := property.(map[string]interface{})
		if !ok {
			return messagesNotFound, errors.New("Expect a map[string]interface{} at path " + path)
		}
		currentPropertyName, ok := p["name"]
		if !ok {
			return messagesNotFound, errors.New("Property name missing at path " + path)
		}
		log.Debug("path=" + path)
		messagesNotFound = append(messagesNotFound, translateProperty(p, first, parentProperty, path, lang)...)
		first = false
		if val, ok := p["properties"]; ok {
			log.Debugf("List of properties found %v", val)
			newProperties, ok := val.([]interface{})
			if !ok {
				return messagesNotFound, errors.New("Expect an []interface{} at path: " + path)
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
			m, err := translateProperties(newProperties, first, p, newPath, lang, messagesNotFound)
			if err != nil {
				return messagesNotFound, err
			}
			messagesNotFound = append(messagesNotFound, m...)
		}
	}
	return messagesNotFound, nil
}

func translateProperty(property map[string]interface{}, first bool, parentProperty map[string]interface{}, path string, lang string) []string {
	log.Debug("Entering in... translateProperty")
	messagesNotFound := make([]string, 0)
	if val, ok := property["description"]; ok {
		property["description"], _ = i18nUtils.Translate(val.(string), val.(string), []string{lang})
		if property["description"].(string) == val.(string) {
			log.Warning("message '" + val.(string) + "' not found")
			messagesNotFound = append(messagesNotFound, val.(string))
		}
	}
	if val, ok := property["label"]; ok {
		property["label"], _ = i18nUtils.Translate(val.(string), val.(string), []string{lang})
		if property["label"].(string) == val.(string) {
			log.Error("message '" + val.(string) + "' not found")
			messagesNotFound = append(messagesNotFound, val.(string))
		}
	}
	if val, ok := property["validation_error_message"]; ok {
		property["validation_error_message"], _ = i18nUtils.Translate(val.(string), val.(string), []string{lang})
		if property["validation_error_message"].(string) == val.(string) {
			log.Warning("message '" + val.(string) + "' not found")
			messagesNotFound = append(messagesNotFound, val.(string))
		}
	}
	if val, ok := property["items"]; ok {
		options := val.([]interface{})
		for index := range options {
			option := options[index].(map[string]interface{})
			optionTranslated, _ := i18nUtils.Translate(option["label"].(string), option["label"].(string), []string{lang})
			if optionTranslated == option["label"].(string) {
				log.Warning("message '" + option["label"].(string) + "' not found")
				messagesNotFound = append(messagesNotFound, option["label"].(string))
			}
			option["label"] = optionTranslated
		}
	}
	if val, ok := property["sample_value"]; ok {
		switch val.(type) {
		case string:
			property["sample_value"], _ = i18nUtils.Translate(val.(string), val.(string), []string{lang})
			if property["sample_value"].(string) == val.(string) {
				log.Warning("message '" + val.(string) + "' not found")
				messagesNotFound = append(messagesNotFound, val.(string))
			}
		default:
		}
	}
	return messagesNotFound
}

func getUIMetadataNameOnly(cfg *config.Config) (*config.Config, error) {
	log.Debug("Entering in... getUIMetadataNameOnly")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	configs, errUIConfigFileCfg := cfg.Map("")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	var names []interface{}
	for key, elem := range configs {
		var name map[string]string
		name = make(map[string]string, 0)
		name["id"] = key
		log.Info(elem)
		if label, ok := (elem.(map[string]interface{}))["label"]; ok {
			name["label"] = label.(string)
		} else {
			name["label"] = key
		}
		names = append(names, name)
	}
	errUIConfigFileCfg = uiConfigFileCfg.Set("ui_metadata", names)
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	return uiConfigFileCfg.Get("ui_metadata")
}
