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
	commandsRunner "github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
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

//
func postInitServer() commandsRunner.InitFunc {
	return commandsRunner.InitFunc(func(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string) {
		//You can add here new handler to enrich the server with new API.
		commandsRunner.AddHandler("/myurl", handlers.HelloWorldHander, false)
		//Specify here parameters for the extensionManager
		extensionManager.Init("examples/data/test-extensions.yml", "examples/extensions", "examples/data/extensions/", "examples/data/logs/extensions")
		//You can overwrite here default value for the configurationManager
		//		configManager.SetConfigFileName("myconfig.yml")
		//		configManager.SetConfigYamlRootKey("myconfig")
	})
}

func main() {
	commandsRunner.ServerStart(nil, postInitServer(), nil)
}
