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
package i18nUtils

import (
	"testing"

	//log "github.com/sirupsen/logrus"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func TestTranslations(t *testing.T) {
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
