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

const DefaultHost = "localhost"
const DefaultPort = "8080"
const DefaultPortSSL = "8483"
const DefaultProtocol = "http"
const DefaultUrl = DefaultProtocol + "://" + DefaultHost + ":" + DefaultPort
const DefaultOutputFormat = "text"
const DefaultTimeout = 60
const DefaultInsecureSSL = false
const SSLCertFileName = "cm-cfp-cert.crt"
const SSLKeyFileName = "cm-cfp-key.pem"
const TokenFileName = "cm-cfp-token"
const CommandsRunnerStatesName = "crs-name"
