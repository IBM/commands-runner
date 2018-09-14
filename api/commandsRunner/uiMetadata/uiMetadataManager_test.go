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
package uiMetadata

import (
	"testing"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
)

const COPYRIGHT_TEST string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func TestGetUIConfigExtentionTest(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//	global.SetExtensionResourcePath("../../test/resource/extensions")
	state.SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	state.SetExtensionPath("../../test/resource/extensions/")
	_, err := GetUIMetaData("ext-template", "test-ui")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetUIConfigError(t *testing.T) {
	_, err := GetUIMetaData("does-not-exist", "test-ui")
	if err == nil {
		t.Error("An error should be raised as this file doesn't exists")
	}
}
