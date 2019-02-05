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
package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HelloWorldHander(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in dummyHander")
	w.Write([]byte("Hello world\n"))
}
