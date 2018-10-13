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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
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
		newPath := filepath.Join(destDir, relPath)
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

//GetHomeDir returns the current user home dir
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

//getExtensionName from request
func GetExtensionNameFromRequest(req *http.Request) (string, url.Values, error) {
	log.Debug("Entering in GetExtensionNameFromRequest")
	m, errRQ := url.ParseQuery(req.URL.RawQuery)
	if errRQ != nil {
		return "", m, errRQ
	}
	var extensionName string
	extensionNameFound, okExtensionName := m["extension-name"]
	if okExtensionName {
		log.Debugf("ExtensionName:%s", extensionNameFound)
		extensionName = extensionNameFound[0]
	} else {
		return "", m, errors.New("extension-name not found in request")
	}
	return extensionName, m, nil
}

func ExtractKey(inputFilePath string, key string) ([]byte, error) {
	log.Debug("Entering in... ExtractKey")
	input, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		return nil, err
	}
	var inYaml map[string]interface{}
	inYaml = make(map[string]interface{}, 0)
	err = yaml.Unmarshal(input, &inYaml)
	if err != nil {
		return nil, err
	}
	var outYaml map[string]interface{}
	outYaml = make(map[string]interface{}, 0)
	outYaml[key] = inYaml[key]
	output, err := yaml.Marshal(outYaml)
	if err != nil {
		return nil, err
	}
	return output, err
}
