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
package uiMetadata

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"

	log "github.com/sirupsen/logrus"

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

func GetUIMetaData(extensionName string, uiMetadataName string) ([]byte, error) {
	log.Debugf("extensionName=%s", extensionName)
	log.Debugf("extensionName=%s", extensionName)
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	if uiMetadataName == "" {
		uiMetadataName = global.DefaultUIMetaDataName
	}
	log.Debugf("uiMetadataName=%s", uiMetadataName)
	raw, e := getUIMetadata(extensionName, uiMetadataName)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIMetadata(extensionName string, uiMetadataName string) ([]byte, error) {
	log.Debug("Entering in... getUIMetadata")
	filePath := filepath.Join(global.ConfigDirectory, "/extension-manifest.yml")
	rootPath := state.GetExtensionPathCustom()
	embeddedExtension, _ := state.IsEmbeddedExtension(extensionName)
	if embeddedExtension {
		rootPath = state.GetExtensionPathEmbedded()
	}
	filePath = filepath.Join(rootPath, extensionName, "/extension-manifest.yml")
	manifest, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	cfg, err := config.ParseYaml(string(manifest))
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get("ui-metadata." + uiMetadataName)
	if err == nil {
		uiConfigFilefg, err := config.ParseYaml("ui-metadata:")
		if err != nil {
			return nil, err
		}
		err = uiConfigFilefg.Set("ui-metadata", cfg.Root)
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
