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
	"path/filepath"
)

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
const CloudFoundryPieName = "pie-cf-deploy"
const DomainTypeEnv = "env"
const DomainTypeApps = "apps"

const CFPWorkload = "CFP"
const VMWareTarget = "vmware"
const OpenStackTarget = "openstack"
const MockTarget = "mock"

const ErrorTypeError = "error"
const ErrorTypeWarning = "warning"

const GeneratedUIConfigPath = "/data/CloudFoundry/uiconfig.yml"

const UIConfigJsonFileName = "uiconfig.json"
const UIConfigYamlFileName = "uiconfig.yml"

const OpenstackAPIVersion = 3
const OpenstackCloudYamlDirPath = ".config" + string(filepath.Separator) + "openstack" + string(filepath.Separator)
