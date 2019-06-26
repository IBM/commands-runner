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
package status

import (
	log "github.com/sirupsen/logrus"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Status struct {
	Name   string `json:"name"`
	Status string `json:"value"`
}

type Statuses map[string]Status

const CMStatus = "cr_status"

var statuses Statuses

//Initialize the properties map
func init() {
	statuses = make(Statuses)
	setInceptionStatus("Initialization")
}

//Retrieve all statuses
func GetStatuses() (*Statuses, error) {
	s, _ := getInceptionStatus()
	statuses[s.Name] = *s
	s, _ = getLogLevel()
	statuses[s.Name] = *s
	return &statuses, nil
}

//Retrieve the inception status
func getInceptionStatus() (*Status, error) {
	var s Status
	s = statuses[CMStatus]
	return &s, nil
}

//Set the inception status
func setInceptionStatus(status string) error {
	SetStatus(CMStatus, status)
	return nil
}

//Set the status
func SetStatus(name string, status string) error {
	var s Status
	s.Name = name
	s.Status = status
	statuses[s.Name] = s
	return nil
}

func getLogLevel() (*Status, error) {
	var s Status
	s.Name = "log_level"
	s.Status = log.GetLevel().String()
	return &s, nil
}
