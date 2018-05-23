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

var UIExtensionResourcePath = "api/resource/extensions/"

//Set the extensionPath
func SetExtensionResourcePath(extensionResourcePathIn string) {
	UIExtensionResourcePath = extensionResourcePathIn
}

var ConfigDirectory string
