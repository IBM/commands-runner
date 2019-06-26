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
package i18nUtils

import (
	"testing"

	//log "github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func Test_i18nUpToDate(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//EN test
	translation, _ := Translate("i18n.test.helloworld", "error", []string{global.DefaultLanguage})
	if translation != "Hello world" {
		t.Error("Expect: 'Hello world' and got '" + translation + "'")
	}
	//FR test
	translation, _ = Translate("i18n.test.helloworld", "error", []string{"fr"})
	if translation != "Bonjour tout le monde" {
		t.Error("Expect: 'Bonjour tout le monde' and got '" + translation + "'")
	}
	//Default test
	translation, _ = Translate("i18n.test.helloworld", "error", []string{""})
	if translation != "Hello world" {
		t.Error("Expect: 'Hello world' and got '" + translation + "'")
	}
	//Bad ID
	translation, _ = Translate("i18n.test.badID", "error", []string{"gr"})
	if translation != "error" {
		t.Error("Expect: 'Hello world' and got '" + translation + "'")
	}
}
