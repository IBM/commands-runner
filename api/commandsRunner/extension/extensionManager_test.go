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
package extension

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestExtensionRegistered(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering... check to see if an extension is registerd")
	SetExtensionPath("../../test/data/extensions/")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	registered := IsExtensionRegistered("cfp-ext-template")
	if registered == false {
		t.Error("extension does not exist")
	}
}

func TestIsIBMExtension(t *testing.T) {
	SetExtensionPath("../../test/data/extensions/")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	isExtension, err := IsEmbeddedExtension("cfp-ext-template")
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

func TestListRegisteredExtensions(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	SetExtensionPath("../../test/data/extensions/")
	extensions, err := ListExtensions("", false)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("%v", extensions)
}
