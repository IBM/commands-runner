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

	"github.com/olebedev/config"
	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/properties"
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
		SetPropertiesEndpoint(w, req)
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
	extensionName, _, err := global.GetExtensionNameFromRequest(req)
	//Retrieve the property name
	property, err := FindProperty(extensionName, params[2])
	if err == nil {
		err = json.NewEncoder(w).Encode(property)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
		}
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
	GetPropertiesEndpoint(w, req)
}

/*
Retrieve all properties
URL: /cr/v1/config/
Method: GET
*/
func GetPropertiesEndpoint(w http.ResponseWriter, req *http.Request) {
	extensionName, m, err := global.GetExtensionNameFromRequest(req)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	uiMetaDataName := global.DefaultUIMetaDataName
	uiMetaDataNameFound, okuiMetaDataName := m["ui-metadata-name"]
	if okuiMetaDataName {
		log.Debugf("uiMetaDataName:%s", uiMetaDataNameFound)
		uiMetaDataName = uiMetaDataNameFound[0]
	}

	//Retrieve properties
	properties, err := GetProperties(extensionName)

	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	properties, err = PropertiesEncodeDecode(extensionName, uiMetaDataName, properties, true)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	cfg, err := config.ParseJson("{}")
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = cfg.Set(global.ConfigRootKey, properties)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := config.RenderJson(cfg.Root)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(result))
	// enc := json.NewEncoder(w)
	// enc.SetIndent("", "  ")
	// err = enc.Encode(bmxConfig)
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
func SetPropertiesEndpoint(w http.ResponseWriter, req *http.Request) {
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
	extensionName, m, errExtName := global.GetExtensionNameFromRequest(req)
	if errExtName != nil {
		logger.AddCallerField().Error(errExtName.Error())
		http.Error(w, errExtName.Error(), 500)
		return
	}
	uiMetaDataName := global.DefaultUIMetaDataName
	uiMetaDataNameFound, okuiMetaDataName := m["ui-metadata-name"]
	if okuiMetaDataName {
		log.Debugf("uiMetaDataName:%s", uiMetaDataNameFound)
		uiMetaDataName = uiMetaDataNameFound[0]
	}
	var cfg *config.Config
	cfg, err = config.ParseJson(string(body))
	if err != nil {
		log.Debug("It is a yaml")
		cfg, err = config.ParseYaml(string(body))
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), 500)
		}
		ps, err = cfg.Map(global.ConfigRootKey)
	} else {
		ps, err = cfg.Map(global.ConfigRootKey)
		ps, _ = PropertiesEncodeDecode(extensionName, uiMetaDataName, ps, false)
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
