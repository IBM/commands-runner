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
package global

import (
	"testing"
)

func TestCsvSplit(t *testing.T) {
	listCsv := "hello, bonjour "
	elems := CSVSplit(listCsv, ",")
	if elems[0] != "hello" {
		t.Error("Expected bonour and got '" + elems[0] + "'")
	}
	if elems[1] != "bonjour" {
		t.Error("Expected bonour and got '" + elems[1] + "'")
	}
}
