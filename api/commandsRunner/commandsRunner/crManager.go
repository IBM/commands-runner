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
package commandsRunner

import (
	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Log struct {
	Level string `yaml:"level" json:"level"`
}

type Settings struct {
	DefaultExtensionName string `yaml:"default_extension_name" json:"default_extension_name"`
	ConfigRootKey        string `yaml:"config_root_key" json:"config_root_key"`
}

var LogPath string

func SetLogPath(logPath string) {
	LogPath = logPath
}

func SetLogLevel(levelRequested string) error {
	level, err := log.ParseLevel(levelRequested)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

//Retrieve level
func GetLogLevel() string {
	return log.GetLevel().String()
}

func SetDefaultExtensionName(defaultExtensionName string) {
	global.DefaultExtensionName = defaultExtensionName
}

//Retrieve Settings
func GetSettings() *Settings {
	settings := &Settings{
		DefaultExtensionName: global.DefaultExtensionName,
		ConfigRootKey:        global.ConfigRootKey,
	}
	return settings
}
