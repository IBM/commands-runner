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

package extensionManager

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/olebedev/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner-test/api/commandsRunner/resourceManager"
)

const IBMExtensions = "IBM"
const CustomExtensions = "custom"

var extensionIBMFile = "api/resource/extensions/ibm-extensions.txt"
var repoLocalPath = "/repo_local/"

var extensionPath = "/data/extensions/"
var extensionDirIBM = "IBM/"
var extensionPathIBM = extensionPath + extensionDirIBM
var extensionDirCustom = "custom/"
var extensionPathCustom = extensionPath + extensionDirCustom

var extensionLogsPath = "/data/logs/extensions/"
var extensionLogsDirIBM = "IBM/"
var extensionLogsPathIBM = extensionLogsPath + extensionLogsDirIBM
var extensionLogsDirCustom = "custom/"
var extensionLogsPathCustom = extensionLogsPath + extensionLogsDirCustom

type Extension struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Extensions struct {
	Extensions map[string]Extension `json:"extensions"`
}

func SetRepoLocalPath(_repoLocalPath string) {
	repoLocalPath = _repoLocalPath
}

//Set the extensionPath
func SetExtensionPath(_extensionPath string) {
	extensionPath = _extensionPath
	extensionPathIBM = _extensionPath + extensionDirIBM
	extensionPathCustom = _extensionPath + extensionDirCustom
}

func GetExtensionPath() string {
	return extensionPath
}

func GetExtensionPathIBM() string {
	return extensionPathIBM
}

func GetExtensionPathCustom() string {
	return extensionPathCustom
}

func GetExtensionLogsPathIBM() string {
	return extensionLogsPathIBM
}

func GetExtensionLogsPathCustom() string {
	return extensionLogsPathCustom
}

func GetRepoLocalPath() string {
	return repoLocalPath
}

func GetRegisteredExtensionPath(extensionName string) (string, error) {
	if !IsExtensionRegistered(extensionName) {
		return "", errors.New(extensionName + " is not registered")
	}
	var extensionPath string
	isIBMExtension, err := IsIBMExtension(extensionName)
	if err != nil {
		return "", err
	}
	if isIBMExtension {
		extensionPath = GetExtensionPathIBM() + extensionName + string(filepath.Separator)
	} else {
		extensionPath = GetExtensionPathCustom() + extensionName + string(filepath.Separator)
	}
	return extensionPath, nil
}

func GetRelativeExtensionPath(extensionName string) string {
	log.Debug("Entering in... GetRelativeExtensionPath")
	var extensionPath string
	//	_, err := resourceManager.Asset(global.UIExtensionResourcePath + "ui-" + extensionName + ".json")
	//Not found means that it is a customer extension
	//	if err != nil {
	isIBMExtension, _ := IsIBMExtension(extensionName)
	log.Debug("isIBMExtension:" + extensionName + " =>" + strconv.FormatBool(isIBMExtension))
	//	if err != nil {
	//		return err
	//	}
	if isIBMExtension {
		extensionPath = extensionDirIBM + extensionName + string(filepath.Separator)
	} else {
		extensionPath = extensionDirCustom + extensionName + string(filepath.Separator)
	}
	return extensionPath
}

//Get extensionPath
func GetRootExtensionPath(rootDir string, extensionName string) string {
	log.Debug("Entering in... GetRootExtensionPath")
	if rootDir == "" {
		return rootDir
	}
	if rootDir[len(rootDir)-1] != filepath.Separator {
		rootDir += string(filepath.Separator)
	}
	extensionPath := rootDir
	if extensionName != global.CloudFoundryPieName && extensionName != "" {
		extensionPath += GetRelativeExtensionPath(extensionName)
		//		extensionPath += relativePath + string(filepath.Separator)
	}
	return extensionPath
}

func SetExtensionIBMFile(_extensionIBMFile string) {
	log.Debug("Entering in... SetExtensionIBMFile")
	extensionIBMFile = _extensionIBMFile
}

//IsExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsExtensionRegistered(extensionName string) bool {
	log.Debug("Entering in... IsExtensionRegistered")
	return extensionName == global.CloudFoundryPieName || IsIBMExtensionRegistered(extensionName) || IsCustomExtensionRegistered(extensionName)
}

//IsCustomExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsCustomExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsCustomExtensionRegistered")
	if _, err := os.Stat(GetExtensionPathCustom() + filename); os.IsNotExist(err) {
		return false
	}
	return true
}

//IsIBMxtensionRegistered Check if an extension is register by browzing the extensions directory
func IsIBMExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsIBMExtensionRegistered")
	if _, err := os.Stat(GetExtensionPathIBM() + filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func Unzip(src, dest, extensionName string) error {
	log.Debug("Entering in... Unzip")
	extensionHomeDir := filepath.Join(dest, extensionName)
	if _, err := os.Stat(extensionHomeDir); os.IsNotExist(err) {
		os.MkdirAll(extensionHomeDir, 0744)
	}
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		err = extractAndWriteFile(dest, extensionName, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractAndWriteFile(targetdir, extensionName string, zf *zip.File) error {
	log.Debug("Entering in... extractAndWriteFile")
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(targetdir, extensionName, zf.Name)
	firstSlash := strings.IndexByte(zf.Name, '/')
	if firstSlash != -1 {
		rootDir := zf.Name[0:firstSlash]
		log.Debug("rootDir:" + rootDir)
		if rootDir == extensionName {
			path = filepath.Join(targetdir, zf.Name)
		}
	}
	log.Debug("Target dir:" + targetdir)
	log.Debug("zf.Name:" + zf.Name)
	log.Debug("path:" + path)
	switch {
	case zf.FileInfo().IsDir():
		os.MkdirAll(path, zf.Mode())

	default:
		os.MkdirAll(filepath.Dir(path), zf.Mode())
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func getIBMExtensionRepoPath(extensionName string) (string, error) {
	log.Debug("Entering in... getIBMExtensionRepoPath")
	log.Debug("extensionName:" + extensionName)
	log.Debug("repoLocalPath:" + repoLocalPath)
	files, err := ioutil.ReadDir(repoLocalPath + extensionName)
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("No version available for IBM extension " + extensionName)
	}
	return repoLocalPath + extensionName + string(filepath.Separator) + files[0].Name() + string(filepath.Separator), nil
}

func CopyExtensionToIBMExtensionPath(extensionName string) error {
	log.Debug("Entering in... CopyExtensionToIBMExtensionPath")
	destDir := GetExtensionPathIBM() + extensionName
	extensionRepoPath, err := getIBMExtensionRepoPath(extensionName)
	if err != nil {
		return err
	}

	extensionRepoPath = strings.TrimRight(extensionRepoPath, string(filepath.Separator))
	log.Debug("extensionRepoPath:" + extensionRepoPath)
	if _, err := os.Stat(extensionRepoPath); err != nil {
		return err
	}

	log.Debug("destDir:" + destDir)
	//Copy every file path to exenstionPathIBM
	visit := func(path string, f os.FileInfo, err error) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		//	newExtensionPath := filepath.Dir(path)
		//	log.Debug("newExtensionPath:" + newExtensionPath)
		log.Debug("path:" + path)
		//newPath := GetExtensionPathIBM() + path[len(repoLocalPath):len(destDir)]
		newPath := strings.Replace(path, extensionRepoPath, GetExtensionPathIBM()+extensionName, 1)
		//newPath := GetExtensionPathIBM() + path[len(extensionRepoPath):] + extensionName
		log.Debug("newPath:" + newPath)
		switch {
		case f.IsDir():
			log.Debug("Create Directory:" + newPath)
			os.MkdirAll(newPath, f.Mode())
		default:
			log.Debug("Create Directory (file not dir):" + filepath.Dir(newPath))
			os.MkdirAll(filepath.Dir(newPath), f.Mode())
			newFile, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer newFile.Close()

			_, err = io.Copy(newFile, file)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := filepath.Walk(extensionRepoPath, visit); err != nil {
		return err
	}
	return nil
}

//IsExtension check if the extensionName is an extension
func IsExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsExtension")
	isIBMExtension, err := IsIBMExtension(extensionName)
	if err != nil {
		return false, err
	}
	isCustomExtension, err := IsCustomExtension(extensionName)
	if err != nil {
		return false, err
	}
	if isIBMExtension || isCustomExtension {
		return true, nil
	}
	return false, nil
}

//IsCustomExtension Check if extensionName is a custom extension
func IsCustomExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsCustomExtension")
	extensions, err := ListRegisteredCustomExtensions()
	if err != nil {
		return false, err
	}
	_, ok := extensions.Extensions[extensionName]
	return ok, nil
}

//IsIBMExtension Check if extensionName is a IBM extension
func IsIBMExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsIBMExtension")
	extensions, err := ListIBMExtensions()
	if err != nil {
		return false, err
	}
	_, ok := extensions.Extensions[extensionName]
	return ok, nil
}

func listRegisteredExtensionsDir(extensionPath string) (*Extensions, error) {
	log.Debug("Entering in... listRegisteredExtensionsDir")
	var extensionList Extensions
	extensionList.Extensions = make(map[string]Extension)
	files, err := ioutil.ReadDir(extensionPath)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			var extension Extension
			extension.Name = file.Name()
			if extensionPath == GetExtensionPathCustom() {
				extension.Type = CustomExtensions
			} else {
				extension.Type = IBMExtensions
			}
			extensionList.Extensions[extension.Name] = extension
		}
	}
	return &extensionList, nil
}

func ListIBMRegisteredExtensions() (*Extensions, error) {
	log.Debug("Entering in... ListIBMRegisteredExtensions")
	extensions, err := listRegisteredExtensionsDir(GetExtensionPathIBM())
	if err != nil {
		return nil, err
	}
	return extensions, nil
}

//ListCustomExtensions returns extensions by reading the custom extension directory
//Custom extensions get be listed only if registered.
func ListRegisteredCustomExtensions() (*Extensions, error) {
	log.Debug("Entering in... ListRegisteredCustomExtensions")
	extensions, err := listRegisteredExtensionsDir(GetExtensionPathCustom())
	if err != nil {
		return nil, err
	}
	return extensions, nil
}

//ListIBMExtensions returns extensions by reading the resourceManager extension file.
func ListIBMExtensions() (*Extensions, error) {
	log.Debug("Entering in... ListIBMExtensions")
	var extensionList Extensions
	extensionList.Extensions = make(map[string]Extension)
	resource, err := resourceManager.Asset(extensionIBMFile)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	resourceArray := bytes.Split(resource, []byte("\n"))

	for _, byteArr := range resourceArray {
		if string(byteArr) != "" {
			var extension Extension
			extension.Name = string(byteArr)
			extension.Type = IBMExtensions
			extensionList.Extensions[extension.Name] = extension
		}
	}
	return &extensionList, nil
}

//ListExtensions list all extensions
func ListExtensions(filter string, catalog bool) (*Extensions, error) {
	log.Debug("Entering in... ListExtensions")
	var extensionList Extensions
	extensionList.Extensions = make(map[string]Extension)
	if catalog {
		extensions, err := ListIBMExtensions()
		if err != nil {
			return nil, err
		}
		extensionList = *extensions
	} else {
		if filter == "" || filter == CustomExtensions {
			extensions, err := ListRegisteredCustomExtensions()
			if err != nil {
				return nil, err
			}
			extensionList = *extensions
		}

		if filter == "" || filter == IBMExtensions {
			var extensions *Extensions
			var err error
			extensions, err = ListIBMRegisteredExtensions()
			if err != nil {
				return nil, err
			}
			//Check if map already contains customer extension
			if len(extensionList.Extensions) == 0 {
				extensionList = *extensions
			} else {
				for k, v := range extensions.Extensions {
					extensionList.Extensions[k] = v
				}
			}
		}
	}
	return &extensionList, nil
}

//Take a backup of an extension on /tmp
func backupExtension(extensionName string) (bool, error) {
	extensionPath, err := GetRegisteredExtensionPath(extensionName)
	if err != nil && extensionPath != "" {
		return false, err
	}
	//extension not yet registered, so no backup needed.
	if extensionPath == "" {
		return false, nil
	}
	backupPath := "/tmp/" + extensionName + string(filepath.Separator)
	err = os.RemoveAll(backupPath)
	if err != nil {
		return false, err
	}
	err = global.CopyRecursive(extensionPath, backupPath)
	log.Debug("Backup of " + extensionPath + " taken at " + backupPath)
	return true, err
}

func restoreExtension(extensionName string) error {
	extensionPath, err := GetRegisteredExtensionPath(extensionName)
	if err != nil {
		return err
	}
	backupPath := "/tmp/" + extensionName + string(filepath.Separator)
	err = os.RemoveAll(extensionPath)
	if err != nil {
		return err
	}
	err = global.CopyRecursive(backupPath, extensionPath)
	log.Debug("Restore of " + extensionPath + " from " + backupPath)
	return err
}

func RegisterExtension(extensionName, zipPath string, force bool) error {
	log.Debug("Entering in... RegisterExtension")
	log.Debug("extensionName: " + extensionName)
	log.Debug("zipPath: " + zipPath)
	if !force && IsExtensionRegistered(extensionName) {
		return errors.New("Extension " + extensionName + " already registered")
	}
	isIBMExtension, err := IsIBMExtension(extensionName)
	log.Debug("isIBMExtension:" + extensionName + " =>" + strconv.FormatBool(isIBMExtension))
	if err != nil {
		return err
	}
	var extensionPath string
	backupTaken, err := backupExtension(extensionName)
	if err != nil {
		return err
	}
	var errInstall error
	if isIBMExtension {
		if zipPath != "" {
			fileInfo, _ := os.Stat(zipPath)
			if fileInfo.Size() != 0 {
				return errors.New("Extension name is already used by IBM extension")
			}
		}
		errInstall = CopyExtensionToIBMExtensionPath(extensionName)
		extensionPath = GetExtensionPathIBM() + extensionName
	} else {
		if zipPath != "" {
			errInstall = Unzip(zipPath, GetExtensionPathCustom(), extensionName)
			if errInstall == nil {
				os.Remove(zipPath)
			}
		} else {
			errInstall = errors.New("the zipPath parameter is missing")
		}
		extensionPath = GetExtensionPathCustom() + extensionName
	}
	var errGen error
	if errInstall == nil {
		errGen = generatePieFile(extensionName, extensionPath)
	}
	if errInstall != nil || errGen != nil {
		if backupTaken {
			restoreExtension(extensionName)
		} else {
			os.RemoveAll(extensionPath)
			os.Remove(extensionPath)
		}
	}
	if errGen != nil {
		return errors.New("Error Generate Pie:" + errGen.Error() + "\nRegistration rolled back")
	}
	if errInstall != nil {
		return errors.New("Error Install:" + errInstall.Error() + "\nRegistration rolled back")
	}
	return nil
}

func generatePieFile(extensionName, extensionPath string) error {
	log.Debug("Entering in... generatePieFile")
	manifest, err := ioutil.ReadFile(extensionPath + string(filepath.Separator) + "extension-manifest.yml")
	if err != nil {
		return err
	}
	cfg, err := config.ParseYaml(string(manifest))
	if err != nil {
		return err
	}
	cfg, err = cfg.Get("states")
	if err != nil {
		return err
	}
	pieCfg, err := config.ParseYaml("states:")
	if err != nil {
		return err
	}
	err = pieCfg.Set("states", cfg.Root)
	if err != nil {
		return err
	}
	out, err := config.RenderYaml(pieCfg.Root)
	if err != nil {
		return err
	}
	log.Debug("pie content:\n" + out)
	return ioutil.WriteFile(extensionPath+string(filepath.Separator)+"pie-"+extensionName+".yml", []byte(out), 0644)

}

//UnregisterExtension delete and extension, deletion of IBM extension is not permitted.
func UnregisterExtension(extensionName string) error {
	log.Debug("Entering in... UnregisterExtension")
	log.Debug("extensionName:", extensionName)
	isIBMExtension, err := IsIBMExtension(extensionName)
	log.Debug("isIBMExtension:" + strconv.FormatBool(isIBMExtension))
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if isIBMExtension {
		return errors.New("Deletion of IBM extension is not permitted")
	}
	log.Debug("IsCustomExtensionRegistered:" + strconv.FormatBool(IsCustomExtensionRegistered(extensionName)))
	if !IsCustomExtensionRegistered(extensionName) {
		return errors.New("This extension is not registered")
	}
	err = os.RemoveAll(GetExtensionPathCustom() + extensionName)
	if err != nil {
		return err
	}
	return nil
}
