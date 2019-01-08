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
	"net/http/httputil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	yaml "gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

const testDir = "../../testFile/"

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
		cleanPath := filepath.Clean(src)
		relPath := path[len(cleanPath):]
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
			reader.Close()
			writer.Close()
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

func ForwardRequest(w http.ResponseWriter, req *http.Request, newURL string) {
	log.Debug("Entering in... ForwardRequest")
	log.Debug("newURL: " + newURL)
	if newURL == "" {
		return
	}
	forwardURL, err := url.Parse(newURL)
	if err != nil {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if forwardURL.Scheme == "" {
		forwardURL.Scheme = "http"
	}
	// req.Header.Add("X-Origin-Host", req.Host)
	if forwardURL.Host == "" {
		forwardURL.Host = "localhost:" + ServerPort
	}
	forwardURL.RawQuery = req.URL.RawQuery
	if forwardURL.Path == req.URL.Path {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, "Calling path is the same as forwared path", http.StatusNotFound)
		return
	}
	// req.Header.Add("X-Forwarded-Host", req.Host)
	log.Debug("forwardURL.String():" + forwardURL.String())
	director := func(req *http.Request) {
		req.URL = forwardURL
	}
	proxy := &httputil.ReverseProxy{Director: director}
	w.Header().Del("Access-Control-Allow-Origin")
	w.Header().Del("Access-Control-Allow-Methods")
	w.Header().Del("Access-Control-Allow-Headers")
	proxy.ServeHTTP(w, req)
}

func CopyToTemp(tempDir string, fileName string) (string, error) {
	err := os.MkdirAll(filepath.Join(testDir, tempDir), 0700)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(fileName)
	if err != nil {
		return "", err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		err = CopyRecursive(fileName, filepath.Join(testDir, tempDir, fi.Name()))
		if err != nil {
			return "", err
		}
		return filepath.Join(testDir, tempDir, fi.Name()), nil
	case mode.IsRegular():
		reader, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
		if err != nil {
			return "", err
		}
		writer, err := os.OpenFile(filepath.Join(testDir, tempDir, fi.Name()), os.O_CREATE|os.O_WRONLY, fi.Mode())
		if err != nil {
			return "", err
		}
		_, err = io.Copy(writer, reader)
		reader.Close()
		writer.Close()
		if err != nil {
			return "", err
		}
		return filepath.Join(testDir, tempDir, fi.Name()), nil
	default:
		return "", errors.New("Can not detect file type " + fileName)
	}
}

func RemoveTemp(fileName string) error {
	log.Debug("fileName:" + filepath.Join(testDir, fileName))
	return os.RemoveAll(filepath.Join(testDir, fileName))
}
