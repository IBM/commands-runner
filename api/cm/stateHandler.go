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
package cm

import (
	"bytes"
	"encoding/json"
	"errors"
	"html"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"

	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/extensionManager"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/global"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/logger"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/stateManager"
	yaml "gopkg.in/yaml.v2"
)

//This need to be removed once extensionHandler done

var pieStatesPath string

var stateManagers map[string]stateManager.States

//initialize the map of stateManagers
func init() {
	stateManagers = make(map[string]stateManager.States)
}

//Set the path of the cf-pie
func SetStatePath(statesPath string) {
	pieStatesPath = statesPath
}

//get the cf-pie path
func getStatePath(extensionName string) string {
	var statesPath string
	if extensionName == global.CloudFoundryPieName || extensionName == "" {
		statesPath = pieStatesPath
	} else {
		statesPath = extensionManager.GetRootExtensionPath(extensionManager.GetExtensionPath(), extensionName)
		statesPath += "pie-" + extensionName + ".yml"
	}
	return statesPath
}

//Add a state manager to the map, directly used only for test method.
func addStateManagerToMap(extensionName string, statesPath string) error {
	log.Debug("Entering in addStateManagerToMap")
	log.Debug("Extension name: " + extensionName)
	sm, err := stateManager.NewStateManager(statesPath)
	if err != nil {
		return err
	}
	stateManagers[extensionName] = *sm
	log.Debug("State Manager added for " + extensionName)
	return nil
}

//Add a stateManager for a given extension
func addStateManager(extensionName string) error {
	log.Debug("Entering in addStateManager")
	log.Debug("Extension name:" + extensionName)
	statesPath := getStatePath(extensionName)
	err := addStateManagerToMap(extensionName, statesPath)
	return err
}

//Remove a stateManager
func removeStateManager(extensionName string) error {
	delete(stateManagers, extensionName)
	return nil
}

//Find a stateManager based on the extensionNAme
func getStateManager(extensionName string) (*stateManager.States, error) {
	if val, ok := stateManagers[extensionName]; ok {
		return &val, nil
	}
	return nil, errors.New("stateManager not found for " + extensionName)
}

//Search for a stateManager and if not found create it
func getAddStateManager(extensionName string) (*stateManager.States, error) {
	log.Debug("Entering in getAddStateManagersss")
	log.Debug("Search for manager: " + extensionName)
	if val, ok := stateManagers[extensionName]; ok {
		log.Debug("Manager already exists, returning it")
		return &val, nil
	}
	log.Debug("Manager already doesn't exist, creating")
	err := addStateManager(extensionName)
	if err != nil {
		return nil, err
	}
	log.Debug("returning creatd manager")
	return getStateManager(extensionName)
}

//getExtensionName from request
func getExtensionNameFromRequest(req *http.Request) (string, url.Values, error) {
	log.Debug("Entering in getExtensionNameFromRequest")
	m, errRQ := url.ParseQuery(req.URL.RawQuery)
	if errRQ != nil {
		return "", m, errRQ
	}
	extensionName := global.CloudFoundryPieName
	extensionNameFound, okExtensionName := m["extension-name"]
	if okExtensionName {
		log.Debugf("ExtensionName:%s", extensionNameFound)
		extensionName = extensionNameFound[0]
	}
	return extensionName, m, nil
}

//Search the stateManager based on the extension-name parameter's request.
func getStateManagerFromRequest(req *http.Request) (*stateManager.States, url.Values, error) {
	log.Debug("Entering in getStateManagerFromRequest")
	log.Debug(req.URL.Path)
	extensionName, m, err := getExtensionNameFromRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("ExtensionName:" + extensionName)
	sm, errSM := getAddStateManager(extensionName)
	if errSM != nil {
		return nil, nil, errSM
	}
	return sm, m, nil
}

//handle States rest api requests
func handleStates(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleStates")
	log.Debugf("handleStates req.URL.Path:%s", req.URL.Path)
	log.Debugf("handleStates RawQuery:%s", req.URL.RawQuery)
	//Check format
	//validatePath := regexp.MustCompile("/cm/v1/states")
	//Retreive the requested state
	m, errRQ := url.ParseQuery(req.URL.RawQuery)
	if errRQ != nil {
		logger.AddCallerField().Error(errRQ.Error())
		http.Error(w, errRQ.Error(), 500)
		return
	}
	log.Debug(m)
	actionFound, okActionFound := m["action"]
	if okActionFound {
		log.Debugf("Action:%s", actionFound)
		action := actionFound[0]
		switch action {
		case "insert":
			switch req.Method {
			case "PUT":
				PutInsertStateStatesEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		case "delete":
			switch req.Method {
			case "PUT":
				PutDeleteStateStatesEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		case "set-statuses":
			switch req.Method {
			case "PUT":
				PutSetStatusesStatesEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		}
	} else {
		switch req.Method {
		case "GET":
			GetStatesEndpoint(w, req)
		case "PUT":
			PutStatesEndpoint(w, req)
		default:
			logger.AddCallerField().Error("Unsupported method:" + req.Method)
			http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
		}
	}
}

/*
Retrieve the states
URL: /cm/v1/states
Method: GET
*/
func GetStatesEndpoint(w http.ResponseWriter, req *http.Request) {
	//Check format
	//validatePath := regexp.MustCompile("/cm/v1/states")
	//Retreive the requested state
	log.Debug(req.URL.Path)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	//Retreive the first-line
	status := ""
	if statusFound, okStatus := m["status"]; okStatus {
		log.Debug("status:%s", statusFound)
		status = statusFound[0]
	}
	states, err := sm.GetStates(status)
	if err == nil {
		//		json.NewEncoder(w).Encode(states)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		err := enc.Encode(states)
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/*
PUT the states
URL: /cm/v1/states?overwrite=<true|false>
Method: GET
*/
func PutStatesEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug(req.URL.Path)
	log.Debugf("RawQuery:%s", req.URL.RawQuery)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		log.Debug(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	/*
		if _, err := os.Stat(sm.StatesPath); os.IsNotExist(err) {
			log.Debug(errors.New("State file " + sm.StatesPath + " doesn't exist"))
			http.Error(w, errSM.Error(), http.StatusBadRequest)
			return
		}
	*/
	overwrite := true
	var errCvt error
	if overwriteFound, okOverwrite := m["overwrite"]; okOverwrite {
		overwrite, errCvt = strconv.ParseBool(overwriteFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert overwrite parameter to boolean "+errCvt.Error(), 500)
			return
		}
	}
	var states stateManager.States
	//log.Debugf("ReqBody:\n%s", req.Body)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	//err := json.NewDecoder(req.Body).Decode(&states)
	//log.Debugf(err.Error())
	if err == nil {
		bodyRaw := buf.String()
		body := html.UnescapeString(bodyRaw)
		err = yaml.Unmarshal([]byte(body), &states)
		if err == nil {
			err = sm.SetStates(states, overwrite)
			if err != nil {
				logger.AddCallerField().Error(err.Error())
				http.Error(w, err.Error(), 500)
			}
		} else {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), 500)
		}
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), 500)
	}
}

/*
PUT insert a state in a state file
URL: /cm/v1/states?extension-name=<extension-name>&action=insert&pos=<int>&before=<bool>
Method: PUT
*/
func PutInsertStateStatesEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering.... PutInsertStateStatesEndpoint")
	log.Debug(req.URL.Path)
	log.Debugf("RawQuery:%s", req.URL.RawQuery)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(sm.StatesPath); os.IsNotExist(err) {
		logger.AddCallerField().Error(errors.New("State file " + sm.StatesPath + " doesn't exist"))
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	var pos int
	var errCvt error
	if posFound, okPos := m["pos"]; okPos {
		pos, errCvt = strconv.Atoi(posFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert first-char parameter to integer "+errCvt.Error(), 500)
			return
		}
	} else {
		http.Error(w, "The position to insert must be provided", http.StatusBadRequest)
		return
	}
	before := true
	if beforeFound, okBefore := m["before"]; okBefore {
		before, errCvt = strconv.ParseBool(beforeFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert before parameter to boolean "+errCvt.Error(), 500)
			return
		}
	}
	stateName := ""
	if stateNameFound, okStateName := m["state-name"]; okStateName {
		stateName = stateNameFound[0]
	}
	var state stateManager.State
	//log.Debugf("ReqBody:\n%s", req.Body)
	buf := new(bytes.Buffer)
	nbBytes, err := buf.ReadFrom(req.Body)
	//err := json.NewDecoder(req.Body).Decode(&states)
	//log.Debugf(err.Error())
	if err == nil {
		log.Debug("Number of bytes in body" + strconv.FormatInt(nbBytes, 10))
		if nbBytes == 0 {
			var insertExtensionName string
			if insertExtensionNameFound, okExtensionName := m["insert-extension-name"]; okExtensionName {
				insertExtensionName = insertExtensionNameFound[0]
			} else {
				http.Error(w, "The extension-name must be provided", http.StatusBadRequest)
				return
			}
			log.Debug("Extension to insert: " + insertExtensionName)
			manifestPath, err := extensionManager.GetRegisteredExtensionPath(insertExtensionName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			manifestPath = manifestPath + string(filepath.Separator) + "extension-manifest.yml"
			log.Debug("manifestPath: " + manifestPath)
			manifestBytes, err := ioutil.ReadFile(manifestPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := config.ParseYaml(string(manifestBytes))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			stateCfg, err := cfg.Get("call_state")
			if err != nil {
				http.Error(w, "extension-manifest.yml: "+err.Error(), http.StatusBadRequest)
				return
			}
			err = stateCfg.Set("name", insertExtensionName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			stateString, err := config.RenderYaml(stateCfg.Root)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Debug("call_state: " + stateString)
			err = yaml.Unmarshal([]byte(stateString), &state)
		} else {
			bodyRaw := buf.String()
			body := html.UnescapeString(bodyRaw)
			err = yaml.Unmarshal([]byte(body), &state)
		}
		if err == nil {
			err = sm.InsertState(state, pos, stateName, before)
			if err != nil {
				logger.AddCallerField().Error(err.Error())
				http.Error(w, err.Error(), 500)
			}
		} else {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), 500)
		}
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), 500)
	}
}

/*
PUT delete a state in a state file
URL: /cm/v1/states?extension-name=<extension-name>&action=delete&pos=<int>
Method: PUT
*/
func PutDeleteStateStatesEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug(req.URL.Path)
	log.Debugf("RawQuery:%s", req.URL.RawQuery)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(sm.StatesPath); os.IsNotExist(err) {
		logger.AddCallerField().Error(errors.New("State file " + sm.StatesPath + " doesn't exist"))
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	var pos int
	var errCvt error
	if posFound, okPos := m["pos"]; okPos {
		pos, errCvt = strconv.Atoi(posFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert first-char parameter to integer "+errCvt.Error(), 500)
		}
	} else {
		http.Error(w, "The position to insert must be provided", http.StatusBadRequest)
	}
	stateName := ""
	if stateNameFound, okStateName := m["state-name"]; okStateName {
		stateName = stateNameFound[0]
	}
	err := sm.DeleteState(pos, stateName)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), 500)
	}
}

/*
PUT set a range of states to a specific status
URL: /cm/v1/states?extension-name=<extension-name>&action=set-statuses&status=<status>&from-state-name=<state_from>&from-inclusive=<bool>&to-state-name=<state_to>&to-inclusive=<bool>
Method: PUT
*/
func PutSetStatusesStatesEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering.... PutSetStatusesStatesEndpoint")
	log.Debug(req.URL.Path)
	log.Debugf("RawQuery:%s", req.URL.RawQuery)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(sm.StatesPath); os.IsNotExist(err) {
		logger.AddCallerField().Error(errors.New("State file " + sm.StatesPath + " doesn't exist"))
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	//Retreive the new status
	statusFound, okStatus := m["status"]
	status := ""
	if okStatus {
		log.Debugf("Status:%s", statusFound)
		status = statusFound[0]
	}
	fromIncluded := false
	if fromIncludedFound, okFromIncluded := m["from-include"]; okFromIncluded {
		var errCvt error
		fromIncluded, errCvt = strconv.ParseBool(fromIncludedFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert from-included parameter to boolean "+errCvt.Error(), 500)
			return
		}
	}
	toIncluded := false
	if toIncludedFound, okToIncluded := m["to-include"]; okToIncluded {
		var errCvt error
		toIncluded, errCvt = strconv.ParseBool(toIncludedFound[0])
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert to-included parameter to boolean "+errCvt.Error(), 500)
			return
		}
	}
	fromStateName := ""
	if fromStateNameFound, okFromStateName := m["from-state-name"]; okFromStateName {
		fromStateName = fromStateNameFound[0]
	}
	toStateName := ""
	if toStateNameFound, okToStateName := m["to-state-name"]; okToStateName {
		toStateName = toStateNameFound[0]
	}
	err := sm.SetStatesStatuses(status, fromStateName, fromIncluded, toStateName, toIncluded)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, "Unable to set statuses:"+err.Error(), 500)
		return
	}
}

//handle State rest api requests
func handleState(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering.... in handleState")
	//Check format
	validatePath := regexp.MustCompile("/cm/v1/state/([^/]+)/([^/]+).*$")
	//Retreive the requested state
	log.Debugf("req.URL.Path:%s", req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	if params == nil {
		switch req.Method {
		case "GET":
			GetStateEndpoint(w, req)
		case "PUT":
			PutStateEndpoint(w, req)
		default:
			logger.AddCallerField().Error("Unsupported method:" + req.Method)
			http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
		}

	} else {
		subCommand := params[2]
		switch subCommand {
		case "log":
			switch req.Method {
			case "GET":
				GetStateLogEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		default:
			logger.AddCallerField().Error("Unknown sub-command:" + subCommand)
			http.Error(w, "Unknown sub-command:"+subCommand, http.StatusBadRequest)
		}
	}
}

/*
Retrieve the state record of a state
URL: /cm/v1/state/<state>
Method: GET
*/
func GetStateEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering.... in GetStateEndpoint")
	//Check format
	validatePath := regexp.MustCompile("/cm/v1/state/(.*)$")
	//Retreive the requested state
	log.Debug(req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debug(params)
	log.Debug(len(params))
	if params == nil || len(params) < 1 {
		logger.AddCallerField().Error("Incorrect request, params missing")
		http.Error(w, "Incorrect request, params missing", http.StatusBadRequest)
	} else {
		//Retrieve the log.
		sm, _, errSM := getStateManagerFromRequest(req)
		if errSM != nil {
			logger.AddCallerField().Error(errSM.Error())
			http.Error(w, errSM.Error(), http.StatusBadRequest)
			return
		}
		state, err := sm.GetState(params[1])
		if err == nil {
			//			json.NewEncoder(w).Encode(state)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			err := enc.Encode(state)
			if err != nil {
				logger.AddCallerField().Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	}
}

/*
Retrieve n lines of a log for a given state starting from line s
URL: /cm/v1/log/state?first-line=s&lines=n
Method: GET
first-line default = 0
lines default = MaxInt64
*/
func GetStateLogEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering.... in GetStateLogEndpoint")
	//Check format
	validatePath := regexp.MustCompile("/cm/v1/state/([^/]+)/log")
	//Retreive the requested state
	log.Debug(req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	//	log.Debugf("Params:%s", params)
	//Parse the query
	//	log.Debugf("RawQuery:%s", req.URL.RawQuery)
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	//Retreive the first-line
	var bychar bool
	var firstPos = int64(1)
	var errCvt error
	if firstLineFound, okFirstLine := m["first-line"]; okFirstLine {
		//		log.Debugf("firstLine:%s", firstLineFound)
		firstPos, errCvt = strconv.ParseInt(firstLineFound[0], 10, 64)
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert first-line parameter to integer "+errCvt.Error(), 500)
		}
		bychar = false
	}
	if firsCharFound, okFirstChar := m["first-char"]; okFirstChar {
		//		log.Debugf("firstChar:%s", firsCharFound)
		firstPos, errCvt = strconv.ParseInt(firsCharFound[0], 10, 64)
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert first-char parameter to integer "+errCvt.Error(), 500)
		}
		bychar = true
	}

	//Retreive the lines
	var length int64
	errCvt = nil
	length = math.MaxInt64
	if lengthFound, okLines := m["length"]; okLines {
		//		log.Debugf("lengthFound:%s", lengthFound)
		length, errCvt = strconv.ParseInt(lengthFound[0], 10, 64)
		if errCvt != nil {
			logger.AddCallerField().Error(errCvt.Error())
			http.Error(w, "Can not convert length parameter to integer "+errCvt.Error(), 500)
		}
	}
	//Retrieve the log.
	logData, err := sm.GetLog(params[1], firstPos, length, bychar)
	if err == nil {
		w.Write(logData)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/*
Set the status for a state
URL: /cm/v1/state/<state>?status=newStatus
Method: PUT
status: newStatus
*/
func PutStateEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering..... PutStateEndpoint")
	//Check format
	validatePath := regexp.MustCompile("/cm/v1/state/([^/]+)")
	//Retreive the requested state
	log.Debug(req.URL.Path)
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debug(params)
	//Parse the query
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	//Retreive the new status
	statusFound, okStatus := m["status"]
	status := ""
	if okStatus {
		log.Debugf("Status:%s", statusFound)
		status = statusFound[0]
	}
	reasonFound, okReason := m["reason"]
	reason := ""
	if okReason {
		reason = reasonFound[0]
	}
	log.Debugf("Reason:%s", reason)
	scriptFound, okScript := m["script"]
	script := ""
	if okScript {
		script = scriptFound[0]
	}
	log.Debugf("Script:%s", script)
	recursivellyFound, okRecursivelly := m["recursivelly"]
	recursivelly := true
	var errRecursivelly error
	if okRecursivelly {
		recursivelly, errRecursivelly = strconv.ParseBool(recursivellyFound[0])
		if errRecursivelly != nil {
			logger.AddCallerField().Error(errRecursivelly.Error())
			http.Error(w, errRecursivelly.Error(), http.StatusBadRequest)
			return
		}
	}
	log.Debugf("recursivelly:%s", strconv.FormatBool(recursivelly))
	scriptTimeoutFound, okScriptTimeout := m["script-timeout"]
	scriptTimeout := -1
	if okScriptTimeout {
		var err error
		scriptTimeout, err = strconv.Atoi(scriptTimeoutFound[0])
		if err != nil {
			logger.AddCallerField().Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	log.Debugf("scriptTimeout:%s", scriptTimeout)
	//Update State
	err := sm.SetState(params[1], status, reason, script, scriptTimeout, recursivelly)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
