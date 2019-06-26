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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/logger"
)

func HandleExtension(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleExtension")
	log.Debugf("req.URL.Path:%s", req.URL.Path)
	log.Debugf("req.Method: %s", req.Method)

	switch req.Method {
	case "POST":
		registerExtension(w, req)
	case "DELETE":
		unregisterExtension(w, req)
	}

}

func HandleExtensions(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleExtensions")
	log.Debugf("req.URL.Path:%s", req.URL.Path)
	log.Debugf("req.Method: %s", req.Method)

	switch req.Method {
	case "GET":
		listExtensions(w, req)
	}
}

func listExtensions(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	filter := ""
	filterFound, okFilter := query["filter"]
	if okFilter {
		log.Debugf("filter:%s", filterFound)
		filter = filterFound[0]
	}
	catalogString := "false"
	catalogFound, okCatalog := query["catalog"]
	if okCatalog {
		log.Debugf("Catalog:%s", catalogFound)
		catalogString = catalogFound[0]
	}
	catalog, err := strconv.ParseBool(catalogString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("Query: %s", filter)
	extensions, err := ListExtensions(filter, catalog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(extensions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func unregisterExtension(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	extensionName := query["extension-name"][0]
	log.Debugf("Query: %s", extensionName)

	err := UnregisterExtension(extensionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func registerExtension(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in... registerExtension")
	extensionName := ""
	m, _ := url.ParseQuery(req.URL.RawQuery)
	if extensionNameFound, okExtensionName := m["extension-name"]; okExtensionName {
		log.Debugf("extensions name :%s", extensionNameFound)
		extensionName = extensionNameFound[0]
	}
	log.Debug("extensionName:" + extensionName)
	forceString := req.Header.Get("Force")
	if forceString == "" {
		forceString = "false"
	}
	log.Debug("forceString:" + forceString)
	force, err := strconv.ParseBool(forceString)
	if err != nil {
		logger.AddCallerField().Errorf("Error converting force to boolean: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	runningToFailedString := req.Header.Get("RunningToFailed")
	if runningToFailedString == "" {
		runningToFailedString = "false"
	}
	log.Debug("runningToFailedtring:" + runningToFailedString)
	runningToFailed, err := strconv.ParseBool(runningToFailedString)
	if err != nil {
		logger.AddCallerField().Errorf("Error converting runningToFailed to boolean: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	file, header, _ := req.FormFile("extension")
	// if err != nil {
	// 	// logger.AddCallerField().Errorf("Unable to parse the form: %v", err)
	// 	// http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	// return
	// }
	if file != nil {
		defer file.Close()
		if extensionName == "" {
			extension := filepath.Ext(header.Filename)
			log.Debug("extension:" + extension)
			extensionName = header.Filename[0 : len(header.Filename)-len(extension)]
		}
	}
	// _, free, _ := global.GetStat("/tmp")
	// if free < uint64(2*header.Size) {
	// 	err = errors.New("Not enough free space to install the extension")
	// 	logger.AddCallerField().Errorf("Error while registring: %v", err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	log.Debug("ExtensionName:" + extensionName)
	//This test is done later too but I added here too to avoid to load the whole extension zip.
	if !force && IsExtensionRegistered(extensionName) {
		err = errors.New("Extension " + extensionName + " already registered")
		logger.AddCallerField().Errorf("Error while registring: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	var zipPath string
	if file != nil {
		out, err := os.Create(filepath.Join("/tmp", extensionName))
		if err != nil {
			logger.AddCallerField().Errorf("Unable to create temp file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		file.Close()
		defer os.Remove(out.Name())
		if err != nil {
			logger.AddCallerField().Errorf("Unable to copy file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		zipPath = out.Name()
	}
	err = RegisterExtension(extensionName, zipPath, force, runningToFailed)
	if err != nil {
		logger.AddCallerField().Errorf("Error while registring: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Extension registration complete"))
}
