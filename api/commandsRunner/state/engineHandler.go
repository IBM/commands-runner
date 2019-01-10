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
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
)

//handle Engine rest api requests
func HandleEngine(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleEngine")
	//Check format
	//validatePath := regexp.MustCompile("/cr/v1/engine.*$")
	//Retreive the requested state
	log.Debug("req.URL.Path:%s", req.URL.Path)
	//params := validatePath.FindStringSubmatch(req.URL.Path)
	m, errRQ := url.ParseQuery(req.URL.RawQuery)
	if errRQ != nil {
		logger.AddCallerField().Error(errRQ.Error())
		http.Error(w, errRQ.Error(), 500)
		return
	}
	//Retreive the new status
	action := FirstState
	actionFound, okActionFound := m["action"]
	if okActionFound {
		log.Debug("Action:%s", actionFound)
		action = actionFound[0]
		switch action {
		case "start":
			switch req.Method {
			case "PUT":
				PutStartEngineEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		case "reset":
			switch req.Method {
			case "PUT":
				PutResetEngineEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		case "reset-execution-info":
			switch req.Method {
			case "PUT":
				PutResetEngineExecutionInfoEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		case "mock":
			switch req.Method {
			case "GET":
				GetMockEndpoint(w, req)
			case "PUT":
				SetMockEndpoint(w, req)
			default:
				logger.AddCallerField().Error("Unsupported method:" + req.Method)
				http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
			}
		}
	} else {
		switch req.Method {
		case "GET":
			GetIsRunningEndpoint(w, req)
		default:
			logger.AddCallerField().Error("Unsupported method:" + req.Method)
			http.Error(w, "Unsupported method:"+req.Method, http.StatusMethodNotAllowed)
		}
	}
}

/*
Start the engine
URL: /cr/v1/egine?action=<action>&from_state=<from_state>&to_state=<to_state>
Method: PUT
action: 'start'
first-state default = first state
to-state default = last staten
*/
func PutStartEngineEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in PutStartEngineEndpoint")
	sm, m, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	running, errRunning := sm.IsRunning()
	if errRunning != nil {
		logger.AddCallerField().Error(errRunning.Error())
		http.Error(w, errRunning.Error(), http.StatusBadRequest)
		return
	}
	if running {
		w.WriteHeader(http.StatusConflict)
		return
	}
	//Retreive the new status
	fromState := FirstState
	fromFound, okFrom := m["from-state"]
	if okFrom {
		log.Debugf("From State:%s", fromFound)
		fromState = fromFound[0]
	}
	toState := LastState
	toFound, okTo := m["to-state"]
	if okTo {
		log.Debugf("To State:%s", toFound)
		toState = toFound[0]
	}
	go sm.Execute(fromState, toState, nil, nil)
}

/*
Reset the engine
URL: /cr/v1/engine?action=<action>
Method: PUT
action: 'reset'
*/
func PutResetEngineEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in PutResetEngineEndpoint")
	sm, _, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	log.Debug(req.URL.Path)
	errReset := sm.ResetEngine()
	if errReset != nil {
		logger.AddCallerField().Error(errReset.Error())
		http.Error(w, errReset.Error(), 500)
		return
	}
}

/*
Reset the engine
URL: /cr/v1/engine?action=<action>
Method: PUT
action: 'reset-execution-info'
*/
func PutResetEngineExecutionInfoEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in PutResetEngineEndpoint")
	sm, _, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		logger.AddCallerField().Error(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	log.Debug(req.URL.Path)
	errReset := sm.ResetEngineExecutionInfo()
	if errReset != nil {
		logger.AddCallerField().Error(errReset.Error())
		http.Error(w, errReset.Error(), 500)
		return
	}
}

/*
Check if engine is running
URL: /cr/v1/engine
Method: GET
*/
func GetIsRunningEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in GetIsRunningEndpoint")
	sm, _, errSM := getStateManagerFromRequest(req)
	if errSM != nil {
		log.Debug(errSM.Error())
		http.Error(w, errSM.Error(), http.StatusBadRequest)
		return
	}
	log.Debug(req.URL.Path)
	running, err := sm.IsRunning()
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	if running {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func GetMockEndpoint(w http.ResponseWriter, req *http.Request) {
	data := GetMock()
	mock := &Mock{
		Mock: data,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(mock)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SetMockEndpoint(w http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	log.Debugf("Query: %s", query)

	mock := ""
	mockFound, okMock := query["mock"]
	if okMock {
		log.Debug("mock:%s", mockFound)
		mock = mockFound[0]
	}
	mockBool, err := strconv.ParseBool(mock)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	SetMock(mockBool)
}
