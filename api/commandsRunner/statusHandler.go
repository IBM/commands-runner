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
package commandsRunner

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/statusManager"
)

//handle Status rest api requests
func handleStatus(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleStatus")
	switch req.Method {
	case "GET":
		getStatusesEndpoint(w, req)
	case "PUT":
		setStatusesEndpoint(w, req)
	default:
		http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
	}
}

/*
Retrieve all Status
URL: /cr/v1/status/
Method: GET
*/
func getStatusesEndpoint(w http.ResponseWriter, req *http.Request) {
	//Check format
	regexp.MustCompile("/cr/v1/(status)$")
	//Retrieve statuses
	log.Debug("Check status")
	statuses, err := statusManager.GetStatuses()
	if err == nil {
		json.NewEncoder(w).Encode(statuses)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/*
Set a Status
URL: /cr/v1/status/
Method: PUT
*/
func setStatusesEndpoint(w http.ResponseWriter, req *http.Request) {
	//Check format
	regexp.MustCompile("/cr/v1/(status)$")
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	statusName := ""
	statusNameFound, okStatusName := query["name"]
	if okStatusName {
		log.Debug("statusName:%s", statusNameFound)
		statusName = statusNameFound[0]
	}
	status := ""
	statusFound, okStatus := query["status"]
	if okStatus {
		log.Debug("status:%s", statusFound)
		status = statusFound[0]
	}
	log.Debug("Set status")
	err := statusManager.SetStatus(statusName, status)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
