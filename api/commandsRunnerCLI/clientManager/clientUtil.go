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
package clientManager

import "github.com/olebedev/config"

//Convert JsonToYaml
func convertJSONToYAML(data string) (string, error) {
	cfg, err := config.ParseJson(data)
	if err != nil {
		return "", err
	}
	out, err := config.RenderYaml(cfg.Root)
	if err != nil {
		return "", err
	}
	return out, nil
}

//Convert to yaml only if the requested format is not json.
func (crc *CommandsRunnerClient) convertJSONOrYAML(data string) (string, error) {
	if crc.OutputFormat == "json" {
		return data, nil
	}
	return convertJSONToYAML(data)
}
