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

package logger

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

//DefaultLogLevel default log level
const DefaultLogLevel = "info"

//AddCallerField Add called name to the log.
func AddCallerField() *logrus.Entry {
	if _, f, line, ok := runtime.Caller(1); ok {
		fa := strings.Split(f, "/")
		caller := fmt.Sprintf("%s:%v", fa[len(fa)-1], line)
		return logrus.WithField("caller", caller)
	}
	return &logrus.Entry{}
}
