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
package package commandsRunner

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

const COPYRIGHT_TEST string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

var statesJson string
var stateJson string

func init() {
	directorlog := "sample-director line 1\nsample-director line 2\nsample-director line 3\nsample-director line 4\nsample-director line 5"
	cflog := "sample-cf line 1\nsample-cf line 2\nsample-cf line 3\nsample-cf line 4\nsample-cf line 5"
	ioutil.WriteFile("/tmp/sample-director.log", []byte(directorlog), 0600)
	errDirector := os.Chmod("/tmp/sample-director.log", 0600)
	if errDirector != nil {
		log.Fatal(errDirector.Error())
	}
	ioutil.WriteFile("/tmp/sample-cf.log", []byte(cflog), 0600)
	errCF := os.Chmod("/tmp/sample-cf.log", 0600)
	if errCF != nil {
		log.Fatal(errCF.Error())
	}
	statesJson = "{\"states\":[{\"name\":\"director\",\"label\":\"Director\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"},{\"name\":\"cf\",\"label\":\"CloudFoundry\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}]}"
	stateJson = "{\"name\":\"cfp-ext-template\",\"label\":\"Insert\",\"status\":\"READY\",\"start_time\":\"\",\"end_time\":\"\",\"reason\":\"\"}"
}
