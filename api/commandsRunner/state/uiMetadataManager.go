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

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
)

func GetUIMetaDataConfig(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... GetUIMetaDataConfig")
	log.Debugf("extensionName=%s", extensionName)
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	if uiMetadataName == "" {
		uiMetadataName = global.DefaultUIMetaDataName
	}
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	raw, e := getUIMetadataConfig(extensionName, uiMetadataName, langs)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadataConfig(extensionName string, uiMetadataName string, langs []string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadataConfig")
	cfg, err := getUIMetadataParseConfig(extensionName, uiMetadataName, langs)
	if err == nil {
		uiConfigFilefg, err := config.ParseYaml("ui_metadata:")
		if err != nil {
			return nil, err
		}
		err = uiConfigFilefg.Set("ui_metadata", cfg.Root)
		if err != nil {
			return nil, err
		}
		out, err := config.RenderJson(uiConfigFilefg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available for " + extensionName + " and " + uiMetadataName)
}

func getUIMetadataParseConfig(extensionName string, uiMetadataName string, langs []string) (cfg *config.Config, err error) {
	log.Debug("Entering in... getUIMetadataParseConfig")
	cfg, err = getUIMetadataParseConfigs(extensionName, langs)
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get(uiMetadataName)
	return cfg, err
}
