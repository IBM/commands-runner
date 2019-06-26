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
package commandsRunner

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestBadLevel(t *testing.T) {
	err := SetLogLevel("badLevel")
	if err == nil {
		t.Error("Expecting error because a wrong level is requested")
	}
}

func TestLevel(t *testing.T) {
	err := SetLogLevel(log.ErrorLevel.String())
	if err != nil {
		t.Error(err.Error())
	}
	level := GetLogLevel()
	if level != log.ErrorLevel.String() {
		t.Error("Expect " + log.ErrorLevel.String() + " and got " + level)
	}
}
func TestSettings(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	// err := i18nUtils.RestoreFiles()
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	// err = i18nUtils.LoadMessageFiles()
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	settings := GetSettings([]string{"jp"})
	if settings.DeploymentName != "Deployment tool" {
		t.Error("Expect: 'Deployment tool' and got " + settings.DeploymentName)
	}
	t.Log("deployment name: " + settings.DeploymentName)
}
