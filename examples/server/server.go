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
package main

import (
	log "github.com/sirupsen/logrus"

	commandsRunner "github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner"
	cr "github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/status"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/server/handlers"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
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
		cr.SetDeploymentName("Simple deployment example")
		cr.SetDefaultExtensionName("simple-embedded-extension-without-version")
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
