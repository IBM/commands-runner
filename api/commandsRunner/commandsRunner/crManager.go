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
package commandsRunner

import (
	log "github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/logger"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"
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
	Level      string `yaml:"level" json:"level"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
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

func SetLogMaxBackups(maxBackups int) {
	logger.InitLogFile(global.ConfigDirectory, maxBackups)
}

//Retrieve maxBackup
func GetLogMaxBackups() int {
	return logger.LogFile.MaxBackups
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
