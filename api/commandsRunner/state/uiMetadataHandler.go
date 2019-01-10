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

//handle BMXCOnfig rest api requests
func HandleUIMetadata(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in HandleUIMetadata")
	switch req.Method {
	case "GET":
		getUIMetadataEndpoint(w, req)
	}
}

/*
Retrieve all Status
URL: /cr/v1/uiconfig/
Method: GET
*/
func getUIMetadataEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in getUIMetadataEndpoint")
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
	uiconfig, err := GetUIMetaDataConfig(extensionName, uiMetaDataName)
	if err == nil {
		//		log.Debug(string(uiconfig))
		w.Write([]byte(uiconfig))
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
