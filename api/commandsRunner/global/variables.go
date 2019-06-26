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

//ConfigDirectory directory where the config file can be found
var ConfigDirectory string

//ConfigYamlFileName default file name for the config file.
var ConfigYamlFileName = "config.yml"

//ConfigRootKey default root key for the yaml config file.
var ConfigRootKey = "config"

//Deployment name, mainly use by the UI
//var DeploymentName string

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

//About URL
var AboutURL string

//Default About
var About string
