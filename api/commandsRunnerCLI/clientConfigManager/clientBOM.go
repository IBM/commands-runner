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
package configManagerClient

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/olebedev/config"
)

//GetBOM returns the BOM
func (cmc *ConfigManagerClient) GetBOM() (string, error) {
	url := "cfp/bom"
	data, errCode, err := cmc.restCall(http.MethodGet, url, nil, nil)
	if err != nil {
		return "", err
	}
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get bom, please check logs")
	}
	//Generate the text format otherwize return the json
	if cmc.OutputFormat == "text" {
		cfg, err := config.ParseJson(data)
		if err != nil {
			return "", err
		}
		version, err := cfg.String("bluemix_bom.version")
		if err != nil {
			return "", err
		}
		out := fmt.Sprintf("version: %s\n", version)
		component, err := cfg.String("bluemix_bom.inception")
		if err != nil {
			return "", err
		}
		out += fmt.Sprintf("%s\n", component)
		volumes, err := cfg.List("bluemix_bom.volume_images")
		if err != nil {
			return "", err
		}
		for volumeIndex := range volumes {
			component, err = cfg.String("bluemix_bom.volume_images." + strconv.Itoa(volumeIndex) + ".image")
			if err != nil {
				return "", err
			}
			out += fmt.Sprintf("%s\n", component)
		}
		return out, nil
	}
	return cmc.convertJSONOrYAML(data)
}
