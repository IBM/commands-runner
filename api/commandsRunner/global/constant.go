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
package global

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

//Server config file name
const CommandsRunnerConfigFileName = "commands-runner.yml"

//DefaultHost default host server
const DefaultHost = "localhost"

//DefaultPort default port
const DefaultPort = "30101"

//DefaultPortSSL default port for SSL connection
const DefaultPortSSL = "30103"

//DefaultProtocol default protocol
const DefaultProtocol = "http"

//DefaultUrl default api url
const DefaultUrl = DefaultProtocol + "://" + DefaultHost + ":" + DefaultPort

//BaseURL base url api
const BaseURL = "/cr/v1/"

//CommandsRunnerLogFileName the logFile for CR
const CommandsRunnerLogFileName = "commands-runner.log"

//DefaultOutputFormat default output format
const DefaultOutputFormat = "text"

//DefaultTimeout default timeout for http request
const DefaultTimeout = 180

//DefaultInsecureSSL by defualt the request are secured.
const DefaultInsecureSSL = false

//SSLCertFileName certificate file name
const SSLCertFileName = "cr-cert.crt"

//SSLKeyFileName key file name
const SSLKeyFileName = "cr-key.pem"

//TokenFileName token file name
const TokenFileName = "cr-token"

//DefaultUIMetaDataName default ui metadata attribute
const DefaultUIMetaDataName = "default"

//StatesFileName the states file name to use accross the different packages
const StatesFileName = "states-file.yml"

//DefaultLanguage
const DefaultLanguage = "en-US"

//DefaultExtenstionManifestFile
const DefaultExtenstionManifestFile = "extension-manifest.yml"
