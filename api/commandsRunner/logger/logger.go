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

package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
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
