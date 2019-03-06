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

package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

//DefaultLogLevel default log level
const DefaultLogLevel = "info"

//Default maximum backups for log
const DefaultLogMaxBackup = "10"

//Log writer
var LogFile *lumberjack.Logger

//AddCallerField Add called name to the log.
func AddCallerField() *logrus.Entry {
	if _, f, line, ok := runtime.Caller(1); ok {
		fa := strings.Split(f, "/")
		caller := fmt.Sprintf("%s:%v", fa[len(fa)-1], line)
		return logrus.WithField("caller", caller)
	}
	return &logrus.Entry{}
}

func InitLogFile(configDir string, maxBackups int) {
	LogFile = &lumberjack.Logger{
		Filename: filepath.Join(configDir, global.CommandsRunnerLogFileName),
		//	MaxSize:    2, // default 100 megabytes
		MaxBackups: maxBackups,
		// MaxAge:     28, // no age days
		// Compress:   true, // disabled by default
	}
}
