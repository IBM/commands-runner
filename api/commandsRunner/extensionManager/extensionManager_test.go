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
package extensionManager

import (
	"testing"
)

func TestExtensionRegistered(t *testing.T) {
	t.Log("Entering... check to see if an extension is registerd")
	registered, err := IsExtensionRegistered("../../test/resource/dummy-extension/", "fake-text.txt")
	if err != nil || registered == false {
		t.Error("extension does not exist")
	}
}

func TestIsIBMExtension(t *testing.T) {
	SetExtensionPath("../../test/data/extensions/")
	isExtension, err := IsIBMExtension("cfp-ext-template")
	if err != nil {
		t.Error(err.Error())
	}
	if !isExtension {
		t.Error("cfp-ext-template is an IBM extension")
	}
}

func TestIsExtension(t *testing.T) {
	SetExtensionPath("../../test/data/extensions/")
	isExtension, err := IsExtension("cfp-ext-template")
	if err != nil {
		t.Error(err.Error())
	}
	if !isExtension {
		t.Error("cfp-ext-template is an IBM extension")
	}
}
