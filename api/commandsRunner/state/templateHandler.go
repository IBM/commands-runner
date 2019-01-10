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

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
)

//handle template rest api requests
func HandleTemplate(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in HandleTemplate")
	switch req.Method {
	case "GET":
		getTemplateEndpoint(w, req)
	}
}

/*
Retrieve template
URL: /cr/v1/template/
Method: GET
*/
func getTemplateEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in getTemplateEndpoint")
	extensionName, m, err := global.GetExtensionNameFromRequest(req)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	log.Debug("ExtensionName:" + extensionName)
	uiMetaDataName := global.DefaultUIMetaDataName
	uiMetaDataNameFound, okuiMetaDataName := m["ui-metadata-name"]
	if okuiMetaDataName {
		log.Debugf("uiMetaDataName:%s", uiMetaDataNameFound)
		uiMetaDataName = uiMetaDataNameFound[0]
	}
	//Retrieve the property name
	uiconfig, err := GenerateUIMetaDataTemplate(extensionName, uiMetaDataName)
	if err == nil {
		//		log.Debug(string(uiconfig))
		w.Write([]byte(uiconfig))
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
