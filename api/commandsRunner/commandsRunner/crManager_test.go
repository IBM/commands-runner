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
