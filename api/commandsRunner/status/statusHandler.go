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
package status

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/logger"
)

//handle Status rest api requests
func HandleStatus(w http.ResponseWriter, req *http.Request) {
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
	statuses, err := GetStatuses()
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
		log.Debugf("statusName:%s", statusNameFound)
		statusName = statusNameFound[0]
	}
	status := ""
	statusFound, okStatus := query["status"]
	if okStatus {
		log.Debugf("status:%s", statusFound)
		status = statusFound[0]
	}
	log.Debug("Set status")
	err := SetStatus(statusName, status)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
