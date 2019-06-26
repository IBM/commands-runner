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
package state

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/logger"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"
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
