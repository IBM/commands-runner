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
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/pcmManager"
)

func handlePCM(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handlePCM")
	validatePath := regexp.MustCompile("/cm/v1/(pcm)/log/([\\w]*)$")
	log.Debug(req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debug(params)
	switch req.Method {
	case "GET":
		if params[2] == "level" {
			GetLogLevelEndpoint(w, req)
		} else {
			http.Error(w, "Unsupported url:"+req.URL.Path, http.StatusBadRequest)
		}
	case "PUT":
		if params[2] == "level" {
			SetLogLevelEndpoint(w, req)
		} else {
			http.Error(w, "Unsupported url:"+req.URL.Path, http.StatusBadRequest)
		}
	default:
		logger.AddCallerField().Error("Unsupported method:" + req.Method)
		http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
	}
}

func GetLogLevelEndpoint(w http.ResponseWriter, req *http.Request) {
	data := pcmManager.GetLogLevel()
	logLevel := &pcmManager.Log{
		Level: data,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(logLevel)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SetLogLevelEndpoint(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	level := ""
	levelFound, okLevel := query["level"]
	if okLevel {
		log.Debug("level:%s", levelFound)
		level = levelFound[0]
	}
	err := pcmManager.SetLogLevel(level)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
