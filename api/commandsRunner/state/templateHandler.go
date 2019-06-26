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

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/logger"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"
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
	langs := i18nUtils.GetLangs(req)
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
	uiconfig, err := GenerateUIMetaDataTemplate(extensionName, uiMetaDataName, langs)
	if err == nil {
		//		log.Debug(string(uiconfig))
		w.Write([]byte(uiconfig))
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
