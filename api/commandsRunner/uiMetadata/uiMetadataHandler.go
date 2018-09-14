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
package uiMetadata

import (
	"net/http"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

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
	//Check format
	validatePath := regexp.MustCompile("/cr/v1/(uimetadata)/([a-z,A-Z,0-9,-]*)/([a-z,A-Z,0-9,-]*)$")
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debugf("params=%s", params)
	log.Debug("params size:" + strconv.Itoa(len(params)))
	// if len(params) < 4 {
	// 	logger.AddCallerField().Error("Configuration name not found")
	// 	http.Error(w, "Configuration name not found", http.StatusBadRequest)
	// 	return
	// }
	extensionName := ""
	if len(params) > 2 {
		extensionName = params[2]
	}
	uiMetadataName := ""
	if len(params) > 2 {
		uiMetadataName = params[3]
	}
	//Retrieve the property name
	config, err := GetUIMetaData(extensionName, uiMetadataName)
	if err == nil {
		w.Write(config)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
