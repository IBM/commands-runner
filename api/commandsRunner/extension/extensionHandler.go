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
package extension

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
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
		listExtension(w, req)
	}
}

func listExtension(w http.ResponseWriter, req *http.Request) {
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
	/*
		if filter == "" {
			io.WriteString(w, `{"extensions":{"extensionsIBM": [`+strings.Join(ListExtensions(GetExtensionPathIBM()), ", ")+`],"extensionsCustom": [`+ strings.Join(ListExtensions(GetExtensionPathCustom()), ", ")+ "]}}")
		} else if strings.ToLower(filter) == "ibm" {
			io.WriteString(w, `{"extensions":{"extensionsIBM": [`+strings.Join(ListExtensions(GetExtensionPathIBM()), ", ")+"]}}")
		} else if strings.ToLower(filter) == "custom" {
			io.WriteString(w, `{"extensions":{"extensionsCustom": [`+strings.Join(ListExtensions(GetExtensionPathCustom()), ", ")+"]}}")
		} else {
			io.WriteString(w, `{"error": "Bad filter query, please use IBM, custom, or leave the parameter emtpy"}`)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	*/
}

func unregisterExtension(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	extensionName := query["name"][0]
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
	//Get filename from zip
	log.Debug("Content-Disposition:" + req.Header.Get("Content-Disposition"))
	mediaType, params, _ := mime.ParseMediaType(req.Header.Get("Content-Disposition"))
	log.Debug("mediaType:" + mediaType)
	filename := params["filename"]
	log.Debug("filename:" + filename)
	extension := filepath.Ext(filename)
	log.Debug("extension:" + extension)
	extensionName := req.Header.Get("Extension-Name")
	if extensionName == "" {
		extensionName = filename[0 : len(filename)-len(extension)]
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
	zipPath := ""
	if filename != "" {
		file, err := ioutil.TempFile("/tmp", extensionName)
		if err != nil {
			logger.AddCallerField().Errorf("Unable to create temp file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		body, err := readBody(req, "extension")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		size, err := io.Copy(file, bytes.NewReader(body))
		//		size, err := io.Copy(file, req.Body)
		if err != nil {
			logger.AddCallerField().Errorf("Unable to copy body in temp file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if filename != "" {
			zipPath = file.Name()
		}
		err = file.Close()
		if err != nil {
			logger.AddCallerField().Errorf("Unable to close the temp file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if size == 0 {
			os.Remove(file.Name())
		}
	}
	err = RegisterExtension(extensionName, zipPath, force)
	if err != nil {
		logger.AddCallerField().Errorf("Error while registring: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check to make sure this is a valid zip file (first 2 or three bytes are special @joaquin)
	//	if extension != ".zip" {
	//		http.Error(w, "Wrong file type", http.StatusBadRequest)
	//		return
	//	}

	//	err = ioutil.WriteFile(GetExtensionPathCustom()+filename, body, 0777)
	//	if err != nil {
	//		http.Error(w, "Cannot create folder for file", http.StatusInternalServerError)
	//		return
	//	}

	//	Unzip(GetExtensionPathCustom()+filename, GetExtensionPathCustom(), extensionName)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Extension registration complete"))
}

func readBody(req *http.Request, fileName string) ([]byte, error) {
	var body []byte
	mediaType, _, _ := mime.ParseMediaType(req.Header.Get("Content-Disposition"))
	if mediaType == "upload" {
		log.Debug("Not a multipart")
		body, err := ioutil.ReadAll(req.Body)
		return body, err
	} else {
		mReader, _ := req.MultipartReader()
		form, err := mReader.ReadForm(100000)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			return body, err
		}
		if fileHeaders, ok := form.File[fileName]; ok {
			log.Debug("Found part named " + fileName)
			for index, fileHeader := range fileHeaders {
				log.Debug(fileHeader.Filename + " part " + strconv.Itoa(index))
				file, err := fileHeader.Open()
				if err != nil {
					logger.AddCallerField().Error(err.Error())
					return body, err
				}
				content := make([]byte, fileHeader.Size)
				file.Read(content)
				body = append(body, content...)
				file.Close()
			}
		}
	}
	return body, nil
}
