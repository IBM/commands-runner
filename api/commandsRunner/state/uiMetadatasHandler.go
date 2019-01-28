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
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/i18n/i18nUtils"
)

//handle BMXCOnfig rest api requests
func HandleUIMetadatas(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in HandleUIMetadatas")
	switch req.Method {
	case "GET":
		getUIMetadatasEndpoint(w, req)
	}
}

/*
Retrieve all Status
URL: /cr/v1/uiconfig/
Method: GET
*/
func getUIMetadatasEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in getUIMetadatasEndpoint")
	langs := i18nUtils.GetLangs(req)
	extensionName, m, err := global.GetExtensionNameFromRequest(req)
	// Don't test because if extension_name not present we will return all configuration names for all extensions.
	// if err != nil {
	// 	logger.AddCallerField().Error(err.Error())
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	return
	// }
	log.Debug("ExtensionName:" + extensionName)
	namesOnly := false
	namesOnlyFound, okNamesOnly := m["names-only"]
	if okNamesOnly {
		log.Debugf("namesOnly:%s", namesOnlyFound)
		var err error
		namesOnly, err = strconv.ParseBool(namesOnlyFound[0])
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}
	//Retrieve the property name
	config, err := GetUIMetaDataConfigs(extensionName, namesOnly, langs)
	if err == nil {
		w.Write(config)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
