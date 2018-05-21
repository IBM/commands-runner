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
package uiConfigManager

import (
	"errors"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/extensionManager"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/resourceManager"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

/**Retrieve the config JSON
The JSON are automatically generated at build time from the yaml files.
The yaml files are located in the api/resource/ directory.
**/
func GetUIConfig(config string) ([]byte, error) {
	log.Debugf("config=%s\n", config)
	if config == "" {
		return nil, errors.New("Config name not specified")
	}
	raw, e := resourceManager.Asset("api/resource/" + config + ".json")
	if e != nil {
		raw, e = resourceManager.Asset("api/resource/extensions/ui-" + config + ".json")
		if e != nil {
			raw, e = resourceManager.Asset("api/test/resource/extensions/ui-" + config + ".json")
			if e != nil {
				raw, e = getCustomUIConfig(config)
				if e != nil {
					return nil, e
				}
			}
		}
	}
	//	log.Debugf("%s\n", raw)
	return raw, nil
}

func getCustomUIConfig(extensionName string) ([]byte, error) {
	log.Debug("Entering in... getCustomUIConfig")
	manifest, err := ioutil.ReadFile(extensionManager.GetExtensionPathCustom() + extensionName + "/extension-manifest.yml")
	if err != nil {
		return nil, err
	}
	cfg, err := config.ParseYaml(string(manifest))
	if err != nil {
		return nil, err
	}
	cfg, err = cfg.Get("uiconfig")
	if err == nil {
		pieCfg, err := config.ParseYaml("uiconfig:")
		if err != nil {
			return nil, err
		}
		err = pieCfg.Set("uiconfig", cfg.Root)
		if err != nil {
			return nil, err
		}
		out, err := config.RenderJson(pieCfg.Root)
		if err != nil {
			return nil, err
		}
		return []byte(out), nil
	}
	return nil, errors.New("No ui configuration available")
}
