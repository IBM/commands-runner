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
package state

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
)

func GetUIMetaDataConfigs(extensionName string, namesOnly bool) ([]byte, error) {
	log.Debug("Entering in... GetUIMetaDataConfig")
	log.Debugf("extensionName=%s", extensionName)
	var raw []byte
	var e error
	if extensionName == "" {
		raw, e = getUIMetadataAllConfigs(namesOnly)
	} else {
		raw, e = getUIMetadataConfigs(extensionName, namesOnly)
	}
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadataAllConfigs(namesOnly bool) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataAllConfigs")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	cfg, err := getUIMetadataAllConfigsCFg(namesOnly)
	if err == nil {
		errUIConfigFileCfg = config.Set(uiConfigFileCfg.Root, "ui_metadata", cfg.Root)
		if errUIConfigFileCfg != nil {
			return nil, err
		}
		out, err := config.RenderJson(uiConfigFileCfg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available")
}

func getUIMetadataConfigs(extensionName string, namesOnly bool) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataConfigs")
	uiConfigFileCfg, errUIConfigFileCfg := config.ParseYaml("ui_metadata:")
	if errUIConfigFileCfg != nil {
		return nil, errUIConfigFileCfg
	}
	cfg, err := getUIMetadataConfigsCfg(extensionName, namesOnly)
	if err == nil && cfg != nil {
		errUIConfigFileCfg = uiConfigFileCfg.Set("ui_metadata", cfg.Root)
		if errUIConfigFileCfg != nil {
			return nil, err
		}
		out, err := config.RenderJson(uiConfigFileCfg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available for " + extensionName)
}

func getUIMetadataAllConfigsCFg(namesOnly bool) (*config.Config, error) {
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
		cfg, err := getUIMetadataConfigsCfg(key, namesOnly)
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

func getUIMetadataConfigsCfg(extensionName string, namesOnly bool) (*config.Config, error) {
	log.Debug("Entering in... getUIMetadataConfigsCfg")
	cfg, err := getUIMetadataParseConfigs(extensionName)
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

func getUIMetadataParseConfigs(extensionName string) (cfg *config.Config, err error) {
	log.Debug("Entering in... getUIMetadataParseConfigs")
	filePath := filepath.Join(global.ConfigDirectory, "/extension-manifest.yml")
	rootPath := GetExtensionsPathCustom()
	embeddedExtension, _ := IsEmbeddedExtension(extensionName)
	if embeddedExtension {
		rootPath = GetExtensionsPathEmbedded()
	}
	filePath = filepath.Join(rootPath, extensionName, "/extension-manifest.yml")
	manifest, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	cfg, err = config.ParseYaml(string(manifest))
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get("ui_metadata")
	return cfg, err
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
