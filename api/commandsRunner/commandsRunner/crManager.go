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
package commandsRunner

import (
	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/i18n/i18nUtils"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Log struct {
	Level string `yaml:"level" json:"level"`
}

type About struct {
	About string `yaml:"about" json:"about"`
}

type Settings struct {
	DeploymentName       string `yaml:"deployment_name" json:"deployment_name"`
	DefaultExtensionName string `yaml:"default_extension_name" json:"default_extension_name"`
	ConfigRootKey        string `yaml:"config_root_key" json:"config_root_key"`
	AboutURL             string `yaml:"about_url" json:"about_url"`
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

// func SetDeploymentName(deploymentName string) {
// 	global.DeploymentName = deploymentName
// }

func SetAboutURL(aboutURL string) {
	global.AboutURL = aboutURL
}

func SetAbout(about string) {
	global.About = about
}

//Retrieve Settings
func GetSettings(langs []string) *Settings {
	deploymentName, _ := i18nUtils.Translate("deployment.name", "Deployment tool", langs)
	settings := &Settings{
		DeploymentName:       deploymentName,
		DefaultExtensionName: global.DefaultExtensionName,
		ConfigRootKey:        global.ConfigRootKey,
		AboutURL:             global.AboutURL,
	}
	return settings
}

func GetAbout() string {
	if global.About == "" {
		return "No information provided"
	}
	return global.About
}
