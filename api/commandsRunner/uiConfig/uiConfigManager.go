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
package uiConfig

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extension"

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

func GetUIConfig(config string) ([]byte, error) {
	log.Debugf("config=%s", config)
	if config == "" {
		return nil, errors.New("Config name not specified")
	}
	raw, e := getUIConfig(config)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

func getUIConfig(extensionName string) ([]byte, error) {
	log.Debug("Entering in... getCustomUIConfig")
	rootPath := extension.GetExtensionPathCustom()
	embeddedExtension, _ := extension.IsEmbeddedExtension(extensionName)
	if embeddedExtension {
		rootPath = extension.GetExtensionPathEmbedded()
	}
	manifest, err := ioutil.ReadFile(filepath.Join(rootPath, extensionName, "/extension-manifest.yml"))
	if err != nil {
		return nil, err
	}
	cfg, err := config.ParseYaml(string(manifest))
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get("uiconfig")
	if err == nil {
		statesFilefg, err := config.ParseYaml("uiconfig:")
		if err != nil {
			return nil, err
		}
		err = statesFilefg.Set("uiconfig", cfg.Root)
		if err != nil {
			return nil, err
		}
		out, err := config.RenderJson(statesFilefg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available")
}
