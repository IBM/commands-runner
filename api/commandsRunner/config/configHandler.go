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
package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/properties"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	yaml "gopkg.in/yaml.v2"
)

//handle COnfig rest api requests
func HandleConfig(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering... handleConfig")
	log.Debug(req.URL.Path)
	log.Debug(req.Method)
	switch req.Method {
	case "GET":
		//search if looking for the all config or a single property
		validatePath := regexp.MustCompile("/cr/v1/(config)$")
		params := validatePath.FindStringSubmatch(req.URL.Path)
		log.Debug(params)
		if len(params) == 2 {
			//Retrieve the full config
			getPropertiesEndpoint(w, req)
		} else {
			//Retrieve a single property
			getPropertyEndpoint(w, req)
		}
	case "POST":
		setPropertiesEndpoint(w, req)
	default:
		http.Error(w, "Unsupported method:"+req.Method, http.StatusNotFound)
	}
}

/*
Retrieve 1 single property
URL: /cr/v1/config/<property_name>
Method: GET
*/
func getPropertyEndpoint(w http.ResponseWriter, req *http.Request) {
	//Check format
	validatePath := regexp.MustCompile("/cr/v1/(config)/([\\w]*)$")
	params := validatePath.FindStringSubmatch(req.URL.Path)
	extensionName, _, err := state.GetExtensionNameFromRequest(req)
	//Retrieve the property name
	property, err := FindProperty(extensionName, params[2])
	if err == nil {
		json.NewEncoder(w).Encode(property)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/*
Retrieve all properties
URL: /cr/v1/config/
Method: GET
*/
func getPropertiesEndpoint(w http.ResponseWriter, req *http.Request) {
	//Check format
	regexp.MustCompile("/cr/v1/(config)$")
	extensionName, _, err := state.GetExtensionNameFromRequest(req)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	//Retrieve properties
	properties, err := GetProperties(extensionName)
	bmxConfig := &Config{
		Properties: properties,
	}

	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	properties, err = PropertiesEncodeDecode(extensionName, properties, true)
	if err != nil {
		log.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(bmxConfig)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*
Set the properties
URL: /cr/v1/config/
Method: POST
*/
func setPropertiesEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering....... setPropertiesEndpoint")
	var ps properties.Properties
	var body []byte
	var err error
	mReader, _ := req.MultipartReader()
	if mReader == nil {
		log.Debug("Not a multipart")
		body, err = ioutil.ReadAll(req.Body)
	} else {
		form, err := mReader.ReadForm(100000)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		if fileHeaders, ok := form.File["config"]; ok {
			log.Debug("Found part named 'config'")
			for index, fileHeader := range fileHeaders {
				log.Debug(fileHeader.Filename + " part " + strconv.Itoa(index))
				file, err := fileHeader.Open()
				if err != nil {
					logger.AddCallerField().Error(err.Error())
					http.Error(w, err.Error(), 500)
					return
				}
				content := make([]byte, fileHeader.Size)
				file.Read(content)
				body = append(body, content...)
				file.Close()
			}
		}
	}
	extensionName, _, errExtName := state.GetExtensionNameFromRequest(req)
	if errExtName != nil {
		logger.AddCallerField().Error(errExtName.Error())
		http.Error(w, errExtName.Error(), 500)
		return
	}
	var bmxConfig Config
	err = json.Unmarshal(body, &bmxConfig)
	//log.Debug("Body:\n" + string(body))
	if err != nil {
		log.Debug("It is a yaml")
		err = yaml.Unmarshal(body, &bmxConfig)
		ps = bmxConfig.Properties
	} else {
		ps, _ = PropertiesEncodeDecode(extensionName, bmxConfig.Properties, false)
	}
	log.Debug("PS decoded")
	if err == nil {
		log.Debug("Set Properties")
		log.Debug("ps len:" + strconv.Itoa(len(ps)))
		err = SetProperties(extensionName, ps)
	}
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), 500)
	}
}
