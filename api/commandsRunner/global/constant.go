/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
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
