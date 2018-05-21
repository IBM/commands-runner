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
	"errors"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
)

//CopyRecursive copy one file (file or dir) to a destDir.
//If the destDir doesn't exist, it will be created.
func CopyRecursive(src, destDir string) error {
	if _, err := os.Stat(src); err != nil {
		log.Debug(err.Error())
		return err
	}
	os.MkdirAll(destDir, 0744)
	err := filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		log.Debug(path)
		relPath := path[len(src)-1:]
		log.Debug("RelPath:" + relPath)
		newPath := destDir + string(filepath.Separator) + relPath
		log.Debug("NewPath:" + newPath)
		if f.IsDir() {
			os.Mkdir(newPath, f.Mode())
		} else {
			reader, err := os.OpenFile(path, os.O_RDONLY, 0666)
			if err != nil {
				return err
			}
			writer, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, f.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, reader)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Debug(err.Error())
		return err
	}
	return nil
}

func CalculateErrorType(currentError string, newError string) string {
	if currentError == ErrorTypeError || newError == ErrorTypeError {
		return ErrorTypeError
	}
	if currentError == ErrorTypeWarning {
		return ErrorTypeWarning
	}
	return newError
}

//Check if a port is accessible for a given IP and protocol
func CheckPort(ip string, protocol string, port int16) error {
	conn, err := net.DialTimeout(protocol, strings.TrimSpace(ip)+":"+strconv.FormatInt(int64(port), 10), time.Duration(5)*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func GetComponentVersion(componentName string) (string, error) {
	log.Debug("Entering... getComponentVersion")
	log.Debug(BOMPath)
	cfgBOM, err := config.ParseYamlFile(BOMPath)
	if err != nil {
		return "", err
	}
	componentsBOM, err := cfgBOM.List("bluemix_bom.volume_images")
	log.Debug(componentsBOM)
	for index := range componentsBOM {
		component, err := cfgBOM.String("bluemix_bom.volume_images." + strconv.Itoa(index) + ".image")
		log.Debug(component)
		if err != nil {
			return "", err
		}
		componentArray := strings.Split(component, ":")
		if componentName == componentArray[0] {
			return componentArray[1], nil
		}
	}
	return "", errors.New("Component " + componentName + " not found")
}

func GetHomeDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		usr, errUsr := user.Current()
		if errUsr != nil {
			log.Debug("Get HOME environment variable")
		} else {
			log.Debug("Use go api to find the home directory")
			homeDir = usr.HomeDir
		}
	}
	log.Debug("HomeDir=" + homeDir)
	return homeDir
}

func CSVSplit(val string, separator string) []string {
	log.Debug("Entering.... CSVSplit")
	elems := strings.Split(val, separator)
	for i := 0; i < len(elems); i++ {
		log.Debug("Trimming: " + elems[i])
		elems[i] = strings.TrimSpace(elems[i])
	}
	log.Debug("Elems:%v", elems)
	return elems
}
