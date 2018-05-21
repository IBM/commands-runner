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
package pcmManager

import (
	log "github.com/sirupsen/logrus"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

type Log struct {
	Level string `yaml:"level" json:"level"`
}

const LogPath = "/data/commands-runner.log"

func SetLogLevel(levelRequested string) error {
	level, err := log.ParseLevel(levelRequested)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

//Retrieve level
func GetLogLevel() string {
	return log.GetLevel().String()
}
