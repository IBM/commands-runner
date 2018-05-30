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

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

//DefaultHost default host server
const DefaultHost = "localhost"

//DefaultPort default port
const DefaultPort = "8080"

//DefaultPortSSL default port for SSL connection
const DefaultPortSSL = "8483"

//DefaultProtocol default protocol
const DefaultProtocol = "http"

//DefaultUrl default api url
const DefaultUrl = DefaultProtocol + "://" + DefaultHost + ":" + DefaultPort

//BaseURL base url api
const BaseURL = "/cr/v1/"

//DefaultOutputFormat default output format
const DefaultOutputFormat = "text"

//DefaultTimeout default timeout for http request
const DefaultTimeout = 60

//DefaultInsecureSSL by defualt the request are secured.
const DefaultInsecureSSL = false

//SSLCertFileName certificate file name
const SSLCertFileName = "cm-cfp-cert.crt"

//SSLKeyFileName key file name
const SSLKeyFileName = "cm-cfp-key.pem"

//TokenFileName token file name
const TokenFileName = "cm-cfp-token"

//CommandsRunnerStatesName default internal state file name
const CommandsRunnerStatesName = "crs-name"
