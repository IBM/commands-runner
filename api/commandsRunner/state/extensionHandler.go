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
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
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
		log.Debug("filter:%s", filterFound)
		filter = filterFound[0]
	}
	catalogString := "false"
	catalogFound, okCatalog := query["catalog"]
	if okCatalog {
		log.Debug("Catalog:%s", catalogFound)
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
		log.Debug("extensions name :%s", extensionNameFound)
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
	file, header, err := req.FormFile("extension")
	if err != nil {
		logger.AddCallerField().Errorf("Unable to parse the form: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// _, free, _ := global.GetStat("/tmp")
	// if free < uint64(2*header.Size) {
	// 	err = errors.New("Not enough free space to install the extension")
	// 	logger.AddCallerField().Errorf("Error while registring: %v", err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	if extensionName == "" {
		extension := filepath.Ext(header.Filename)
		log.Debug("extension:" + extension)
		extensionName = header.Filename[0 : len(header.Filename)-len(extension)]
	}
	log.Debug("ExtensionName:" + extensionName)
	//This test is done later too but I added here too to avoid to load the whole extension zip.
	if !force && IsExtensionRegistered(extensionName) {
		err = errors.New("Extension " + extensionName + " already registered")
		logger.AddCallerField().Errorf("Error while registring: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	out, err := ioutil.TempFile("/tmp", extensionName)
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
	zipPath := out.Name()
	err = RegisterExtension(extensionName, zipPath, force)
	if err != nil {
		logger.AddCallerField().Errorf("Error while registring: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Extension registration complete"))
}
