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
	log "github.com/sirupsen/logrus"

	commandsRunner "github.com/IBM/commands-runner/api/commandsRunner"
	cr "github.com/IBM/commands-runner/api/commandsRunner/commandsRunner"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
	"github.com/IBM/commands-runner/api/commandsRunner/status"
	"github.com/IBM/commands-runner/examples/server/handlers"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func preInitServer() commandsRunner.InitFunc {
	log.Info("preInitServer")
	return commandsRunner.InitFunc(func(port string, portSSL string, configDir string, certificatePath string, keyPath string) {
	})
}

//
func postInitServer() commandsRunner.InitFunc {
	log.Info("postInitServer")
	return commandsRunner.InitFunc(func(port string, portSSL string, configDir string, certificatePath string, keyPath string) {
		//You can add here new handler to enrich the server with new API.
		commandsRunner.AddHandler("/myurl", handlers.HelloWorldHander, false)
		//Specify here parameters for the extensionManager
		state.InitExtensions("examples/extensions/test-extensions.yml", "examples/extensions", "examples/data/extensions/", "examples/data/logs/extensions")
		//You can overwrite here default value for the config package
		//		config.SetConfigFileName("myconfig.yml")
		//		config.SetConfigRootKey("myconfig")
		//The provided value is used when an extension is inserted in the state file
		//in order to call the commands-runner to execute that extension.
		//      config.SetClientPath("./cr-cli")
		cr.SetDefaultExtensionName("default-extension")
		//Specify here the About text, cr.SetAboutURL() can be also use if the server has an API to return the about content.
		cr.SetAbout("This is an example")
	})
}

func preStartServer() commandsRunner.InitFunc {
	log.Info("preStartServer")
	return commandsRunner.InitFunc(func(port string, portSSL string, configDir string, certificatePath string, keyPath string) {
	})
}

func postStartServer() commandsRunner.PostStartFunc {
	log.Info("postStartServer")
	return commandsRunner.PostStartFunc(func(configDir string) {
		//This to tell the config-manager-ui that the installation is completed.
		status.SetStatus("cr_post_install_status", "COMPLETED")
	})
}

func main() {
	commandsRunner.ServerStart(preInitServer(), postInitServer(), preStartServer(), postStartServer())
}
