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

//ConfigDirectory directory where the config file can be found
var ConfigDirectory string

//ConfigYamlFileName default file name for the config file.
var ConfigYamlFileName = "config.yml"

//ConfigRootKey default root key for the yaml config file.
var ConfigRootKey = "config"

//DefaultExtensionName default extension name.
var DefaultExtensionName string

//Mock default false, when true all scripts are skipped
var Mock = false

//Server Configuration director
var ServerConfigDir string

//Server Port
var ServerPort string

//Server Port SSL
var ServerPortSSL string

//Server Certificate path
var ServerCertificatePath string

//Server Key Path
var ServerKeyPath string
