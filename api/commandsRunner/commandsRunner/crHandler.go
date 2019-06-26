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
package commandsRunner

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/logger"
)

func HandleCR(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleCR")
	validatePath := regexp.MustCompile("/cr/v1/cr/(\\blog\\b|\\bsettings\\b|\\babout\\b)(/(\\w[\\w-]*))?$")
	log.Debug(req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debug(params)
	switch params[1] {
	case "log":
		switch req.Method {
		case "GET":
			switch params[3] {
			case "level":
				GetLogLevelEndpoint(w, req)
			case "max-backups":
				GetLogMaxBackupsEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported command:" + params[2])
				http.Error(w, "Unsupported command:"+params[2], http.StatusBadRequest)
			}
		case "PUT":
			switch params[3] {
			case "level":
				SetLogLevelEndpoint(w, req)
			case "max-backups":
				SetLogMaxBackupEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported command:" + params[3])
				http.Error(w, "Unsupported url:"+req.URL.Path, http.StatusBadRequest)
			}
		default:
			logger.AddCallerField().Error("Unsupported method:" + req.Method)
			http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
		}
	case "settings":
		GetSettingsEndpoint(w, req)
	case "about":
		GetAboutEndpoint(w, req)
	default:
		logger.AddCallerField().Error("Unsupported command:" + params[1])
		http.Error(w, "Unsupported command:"+params[1], http.StatusMethodNotAllowed)
	}
}

func GetLogLevelEndpoint(w http.ResponseWriter, req *http.Request) {
	data := GetLogLevel()
	logLevel := &Log{
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

func GetLogMaxBackupsEndpoint(w http.ResponseWriter, req *http.Request) {
	data := GetLogMaxBackups()
	logMaxBackups := &Log{
		MaxBackups: data,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(logMaxBackups)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetSettingsEndpoint(w http.ResponseWriter, req *http.Request) {
	langs := i18nUtils.GetLangs(req)
	settings := GetSettings(langs)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(settings)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetAboutEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in... GetAboutEndpoint")
	log.Debug("global.AboutURL: " + global.AboutURL)
	if global.AboutURL == "" {
		data := GetAbout()
		about := &About{
			About: data,
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		err := enc.Encode(about)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		global.ForwardRequest(w, req, global.AboutURL)
	}
}

func SetLogLevelEndpoint(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	level := ""
	levelFound, okLevel := query["level"]
	if okLevel {
		log.Debugf("level:%s", levelFound)
		level = levelFound[0]
	}
	err := SetLogLevel(level)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SetLogMaxBackupEndpoint(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	logMaxBackups := ""
	logMaxBackupsFound, okLogMaxBackups := query["max-backups"]
	if okLogMaxBackups {
		log.Debugf("max-backups:%s", logMaxBackupsFound)
		logMaxBackups = logMaxBackupsFound[0]
	}
	maxBackups, errMaxBackup := strconv.Atoi(logMaxBackups)
	if errMaxBackup != nil {
		logger.AddCallerField().Error(errMaxBackup.Error())
		http.Error(w, errMaxBackup.Error(), http.StatusInternalServerError)
	}
	SetLogMaxBackups(maxBackups)
}
