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
package uiConfig

import (
	"net/http"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
)

//handle BMXCOnfig rest api requests
func HandleUIConfig(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleUIConfig")
	switch req.Method {
	case "GET":
		getUIConfigEndpoint(w, req)
	}
}

/*
Retrieve all Status
URL: /cr/v1/uiconfig/
Method: GET
*/
func getUIConfigEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in getUIConfigEndpoint")
	//Check format
	validatePath := regexp.MustCompile("/cr/v1/(uiconfig)/([a-z,A-Z,0-9,-]*)$")
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debugf("params=%s", params)
	log.Debug("params size:" + strconv.Itoa(len(params)))
	if len(params) < 3 {
		logger.AddCallerField().Error("Configuration name not found")
		http.Error(w, "Configuration name not found", http.StatusBadRequest)
		return
	}
	//Retrieve the property name
	config, err := GetUIConfig(params[2])
	if err == nil {
		w.Write(config)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
