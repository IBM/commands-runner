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
package statusManager

import (
	log "github.com/sirupsen/logrus"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
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

const CMStatus = "cm_status"

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
