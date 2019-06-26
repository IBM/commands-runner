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
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/commands-runner/api/commandsRunner/logger"
)

//handle Engine rest api requests
func HandleEngine(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleEngine")
	//Check format
	//validatePath := regexp.MustCompile("/cr/v1/engine.*$")
	//Retreive the requested state
	log.Debugf("req.URL.Path: %s", req.URL.Path)
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
		log.Debugf("Action: %s", actionFound)
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
		logger.AddCallerField().Error("Engine Running")
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
	timeNow := time.Now().UTC()
	time.Sleep(1 * time.Second)
	go sm.Execute(fromState, toState, nil, nil)
	for {
		if sm.StartTime != "" {
			timeStart, errTime := time.Parse(time.UnixDate, sm.StartTime)
			if errTime != nil {
				logger.AddCallerField().Error(errTime.Error())
				http.Error(w, errTime.Error(), http.StatusBadRequest)
				return
			}
			log.Debug("Waiting sm.StartTime:" + sm.StartTime)
			log.Debug("Waiting timeNow:" + timeNow.Format(time.UnixDate))
			if timeStart.After(timeNow) {
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
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
		log.Debugf("mock:%s", mockFound)
		mock = mockFound[0]
	}
	mockBool, err := strconv.ParseBool(mock)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	SetMock(mockBool)
}
