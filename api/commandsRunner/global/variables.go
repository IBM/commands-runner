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

//SetExtensionResourcePath the extensionPath
func SetExtensionResourcePath(extensionResourcePathIn string) {
	UIExtensionResourcePath = extensionResourcePathIn
}

//ConfigDirectory directory where the config file can be found
var ConfigDirectory string

//ConfigYamlFileName default file name for the config file.
var ConfigYamlFileName = "config.yml"

//ConfigRootKey default root key for the yaml config file.
var ConfigRootKey = "config"
