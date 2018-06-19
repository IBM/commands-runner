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
	"bufio"
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
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

const EmbeddedExtensions = "embedded"
const CustomExtensions = "custom"

var extensionEmbeddedFile string
var embeddedExtensionsRepositoryPath string

var extensionPath string
var extensionDirEmbedded = EmbeddedExtensions
var extensionPathEmbedded = extensionPath + extensionDirEmbedded
var extensionDirCustom = CustomExtensions + string(filepath.Separator)
var extensionPathCustom = extensionPath + extensionDirCustom

var extensionLogsPath string
var extensionLogsDirEmbedded = extensionDirEmbedded
var extensionLogsPathEmbedded = extensionLogsPath + extensionLogsDirEmbedded
var extensionLogsDirCustom = CustomExtensions
var extensionLogsPathCustom = extensionLogsPath + extensionLogsDirCustom

/*
Extension structure
*/
type Extension struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

/*
Extensions is a map of Extension
*/
type Extensions struct {
	Extensions map[string]Extension `json:"extensions"`
}

/*
Init initialise the extensionManager package.
embeddedExtensionDescriptor the path to the file listing the embedded extension names.
embeddedExtensionsRepositoryPath the directory path where the extensions are located. This path will be extended with "embeded" or "custom" depending of the type of extension.
extensionPath the directory path where the extension will be copied by the registration process. This path will be extended with "embeded" or "custom" depending of the type of extension.
extensionLogsPath the directory path where the logs will be save. This path will be extended with "embeded" or "custom" depending of the type of extension.
*/
func Init(embeddedExtensionDescriptor string, embeddedExtensionsRepositoryPath string, extensionPath string, extensionLogsPath string) {
	SetExtensionEmbeddedFile(embeddedExtensionDescriptor)
	err := SetEmbeddedExtensionsRepositoryPath(embeddedExtensionsRepositoryPath)
	if err != nil {
		log.Fatal(err)
	}
	SetExtensionPath(extensionPath)
	SetExtensionLogsPath(extensionLogsPath)
	err = registerEmbededExtensions(true)
	if err != nil {
		log.Fatal(err)
	}
}

//SetExtensionEmbeddedFile sets the embedded extension path file descriptor
func SetExtensionEmbeddedFile(_extensionEmbeddedFile string) {
	log.Debug("Entering in... SetExtensionEmbeddedFile")
	extensionEmbeddedFile = strings.TrimRight(_extensionEmbeddedFile, string(filepath.Separator))
}

/*
SetEmbeddedExtensionsRepositoryPath set the path where the embedded extension are stored
*/
func SetEmbeddedExtensionsRepositoryPath(_embeddedExtensionsRepositoryPath string) error {
	if _, err := os.Stat(_embeddedExtensionsRepositoryPath); os.IsNotExist(err) {
		return err
	}
	embeddedExtensionsRepositoryPath = strings.TrimRight(_embeddedExtensionsRepositoryPath, string(filepath.Separator))
	return nil
}

//SetExtensionPath set the path where the extensions must be deployed
func SetExtensionPath(_extensionPath string) {
	extensionPath = strings.TrimRight(_extensionPath, string(filepath.Separator))
	extensionPathEmbedded = filepath.Join(extensionPath, extensionDirEmbedded)
	if _, err := os.Stat(extensionPathEmbedded); os.IsNotExist(err) {
		log.Debug("Create dir:" + extensionPathEmbedded)
		os.MkdirAll(extensionPathEmbedded, 0744)
	}
	extensionPathCustom = filepath.Join(extensionPath, extensionDirCustom)
	if _, err := os.Stat(extensionPathCustom); os.IsNotExist(err) {
		log.Debug("Create dir:" + extensionPathCustom)
		os.MkdirAll(extensionPathCustom, 0744)
	}
}

//SetExtensionLogsPath set the path where the extensions logs are kept
func SetExtensionLogsPath(_extensionLogsPath string) {
	extensionLogsPath = strings.TrimRight(_extensionLogsPath, string(filepath.Separator))
	extensionLogsPathEmbedded = filepath.Join(extensionLogsPath, extensionDirEmbedded)
	extensionLogsPathCustom = filepath.Join(extensionLogsPath, extensionDirCustom)
}

//GetExtensionPath retrieves the extension path
func GetExtensionPath() string {
	return extensionPath
}

//GetExtensionPathEmbedded retrieves the extension path for embedded extensions
func GetExtensionPathEmbedded() string {
	return extensionPathEmbedded
}

//GetExtensionPathCustom retrieves the extension path for the custom extensions
func GetExtensionPathCustom() string {
	return extensionPathCustom
}

//GetExtensionLogsPathEmbedded retrieves the embedded extensions logs path
func GetExtensionLogsPathEmbedded() string {
	return extensionLogsPathEmbedded
}

//GetExtensionLogsPathCustom retrieves the custom extensions logs path
func GetExtensionLogsPathCustom() string {
	return extensionLogsPathCustom
}

//GetRepoLocalPath retrieves the location of the embedded extensions packages
func GetRepoLocalPath() string {
	return embeddedExtensionsRepositoryPath
}

//GetRegisteredExtensionPath gets the extension path for a given registered extension
func GetRegisteredExtensionPath(extensionName string) (string, error) {
	if !IsExtensionRegistered(extensionName) {
		return "", errors.New(extensionName + " is not registered")
	}
	var extensionPath string
	isEmbeddedExtension, err := IsEmbeddedExtension(extensionName)
	if err != nil {
		return "", err
	}
	if isEmbeddedExtension {
		extensionPath = filepath.Join(GetExtensionPathEmbedded(), extensionName)
	} else {
		extensionPath = filepath.Join(GetExtensionPathCustom(), extensionName)
	}
	return extensionPath, nil
}

//GetRelativeExtensionPath gets the relative extension path for a given registered extension
func GetRelativeExtensionPath(extensionName string) string {
	log.Debug("Entering in... GetRelativeExtensionPath")
	var extensionPath string
	isEmbeddedExtension, _ := IsEmbeddedExtension(extensionName)
	log.Debug("isEmbeddedExtension:" + extensionName + " =>" + strconv.FormatBool(isEmbeddedExtension))
	if isEmbeddedExtension {
		extensionPath = filepath.Join(extensionDirEmbedded, extensionName)
	} else {
		extensionPath = filepath.Join(extensionDirCustom, extensionName)
	}
	return extensionPath
}

//GetRootExtensionPath gets the root extension path
func GetRootExtensionPath(rootDir string, extensionName string) string {
	log.Debug("Entering in... GetRootExtensionPath")
	if rootDir == "" {
		return rootDir
	}
	if rootDir[len(rootDir)-1] != filepath.Separator {
		rootDir += string(filepath.Separator)
	}
	extensionPath := rootDir
	if extensionName != global.CommandsRunnerStatesName && extensionName != "" {
		extensionPath += GetRelativeExtensionPath(extensionName)
	}
	return extensionPath
}

//IsExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsExtensionRegistered(extensionName string) bool {
	log.Debug("Entering in... IsExtensionRegistered")
	return extensionName == global.CommandsRunnerStatesName || IsEmbeddedExtensionRegistered(extensionName) || IsCustomExtensionRegistered(extensionName)
}

//IsCustomExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsCustomExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsCustomExtensionRegistered")
	if _, err := os.Stat(filepath.Join(GetExtensionPathCustom(), filename)); os.IsNotExist(err) {
		return false
	}
	return true
}

//IsEmbeddedxtensionRegistered Check if an extension is register by browzing the extensions directory
func IsEmbeddedExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsEmbeddedExtensionRegistered")
	log.Debug(filepath.Join(GetExtensionPathEmbedded(), filename))
	if _, err := os.Stat(filepath.Join(GetExtensionPathEmbedded(), filename)); os.IsNotExist(err) {
		return false
	}
	return true
}

//Unzip unzip an extension zip file to a destination directory
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

func getEmbeddedExtensionRepoPath(extensionName string) (string, error) {
	log.Debug("Entering in... getEmbeddedExtensionRepoPath")
	log.Debug("extensionName:" + extensionName)
	log.Debug("embeddedExtensionsRepositoryPath:" + embeddedExtensionsRepositoryPath)

	files, err := ioutil.ReadDir(filepath.Join(embeddedExtensionsRepositoryPath, extensionName))
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("extension directory " + extensionName + " is empty")
	}
	if len(files) == 1 && files[0].IsDir() {
		return filepath.Join(embeddedExtensionsRepositoryPath, extensionName, files[0].Name()), nil
	}
	return filepath.Join(embeddedExtensionsRepositoryPath, extensionName), nil
}

//CopyExtensionToEmbeddedExtensionPath copy the extension to the extension directory
func CopyExtensionToEmbeddedExtensionPath(extensionName string) error {
	log.Debug("Entering in... CopyExtensionToEmbeddedExtensionPath")
	destDir := filepath.Join(GetExtensionPathEmbedded(), extensionName)
	extensionRepoPath, err := getEmbeddedExtensionRepoPath(extensionName)
	if err != nil {
		return err
	}

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
		log.Debug("path:" + path)
		log.Debug("extensionRepoPath:" + extensionRepoPath)
		log.Debug("GetExtensionPathEmbedded()+extensionName:" + filepath.Join(GetExtensionPathEmbedded(), extensionName))
		newPath := strings.Replace(path, extensionRepoPath, filepath.Join(GetExtensionPathEmbedded(), extensionName), 1)
		log.Debug("newPath:" + newPath)
		switch {
		case f.IsDir():
			log.Debug("Create Directory:" + newPath)
			err = os.MkdirAll(newPath, f.Mode())
			if err != nil {
				log.Error(err.Error())
				return err
			}
		default:
			log.Debug("Create Directory (file not dir):" + filepath.Dir(newPath))
			err := os.MkdirAll(filepath.Dir(newPath), f.Mode())
			if err != nil {
				log.Error(err.Error())
				return err
			}
			newFile, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				log.Error(err.Error())
				return err
			}
			defer newFile.Close()

			_, err = io.Copy(newFile, file)
			if err != nil {
				log.Error(err.Error())
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

//IsExtension checks if the extensionName is an extension
func IsExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsExtension")
	isEmbeddedExtension, err := IsEmbeddedExtension(extensionName)
	if err != nil {
		return false, err
	}
	isCustomExtension, err := IsCustomExtension(extensionName)
	if err != nil {
		return false, err
	}
	if isEmbeddedExtension || isCustomExtension {
		return true, nil
	}
	return false, nil
}

//IsCustomExtension Checks if extensionName is a custom extension
func IsCustomExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsCustomExtension")
	extensions, err := ListRegisteredCustomExtensions()
	if err != nil {
		return false, err
	}
	_, ok := extensions.Extensions[extensionName]
	return ok, nil
}

//IsEmbeddedExtension Checks if extensionName is a Embedded extension
func IsEmbeddedExtension(extensionName string) (bool, error) {
	log.Debug("Entering in... IsEmbeddedExtension")
	extensions, err := ListEmbeddedExtensions()
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
		return &extensionList, nil
	}
	for _, file := range files {
		if file.IsDir() {
			var extension Extension
			extension.Name = file.Name()
			if extensionPath == GetExtensionPathCustom() {
				extension.Type = CustomExtensions
			} else {
				extension.Type = EmbeddedExtensions
			}
			extensionList.Extensions[extension.Name] = extension
		}
	}
	return &extensionList, nil
}

//ListEmbeddedRegisteredExtensions lists the registered embedded exxtensions
func ListEmbeddedRegisteredExtensions() (*Extensions, error) {
	log.Debug("Entering in... ListEmbeddedRegisteredExtensions")
	extensions, err := listRegisteredExtensionsDir(GetExtensionPathEmbedded())
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

//ListEmbeddedExtensions returns extensions by reading the resourceManager extension file.
func ListEmbeddedExtensions() (*Extensions, error) {
	log.Debug("Entering in... ListEmbeddedExtensions")
	var extensionList Extensions
	extensionList.Extensions = make(map[string]Extension)
	log.Debug("extensionEmbeddedFile:" + extensionEmbeddedFile)
	if extensionEmbeddedFile == "" {
		return &extensionList, nil
	}
	resource, err := ioutil.ReadFile(extensionEmbeddedFile)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	//resource, err := resourceManager.Asset(extensionEmbeddedFile)

	log.Debug("Extension file content:\n" + string(resource))
	resourceArray := bytes.Split(resource, []byte("\n"))

	for _, byteArr := range resourceArray {
		log.Debug("byteArr:" + string(byteArr))
		if string(byteArr) != "" {
			var extension Extension
			extension.Name = string(byteArr)
			extension.Type = EmbeddedExtensions
			extensionList.Extensions[extension.Name] = extension
		}
	}

	return &extensionList, nil
}

//ListExtensions lists all extensions
func ListExtensions(filter string, catalog bool) (*Extensions, error) {
	log.Debug("Entering in... ListExtensions")
	var extensionList Extensions
	extensionList.Extensions = make(map[string]Extension)
	if catalog {
		extensions, err := ListEmbeddedExtensions()
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

		if filter == "" || filter == EmbeddedExtensions {
			var extensions *Extensions
			var err error
			extensions, err = ListEmbeddedRegisteredExtensions()
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

//registerEmbededExtension register all embeded extensions
func registerEmbededExtensions(force bool) error {
	file, err := os.Open(extensionEmbeddedFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		RegisterExtension(scanner.Text(), "", force)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

//RegisterExtension register an extension
func RegisterExtension(extensionName, zipPath string, force bool) error {
	log.Debug("Entering in... RegisterExtension")
	log.Debug("extensionName: " + extensionName)
	log.Debug("zipPath: " + zipPath)
	if !force && IsExtensionRegistered(extensionName) {
		return errors.New("Extension " + extensionName + " already registered")
	}
	isEmbeddedExtension, err := IsEmbeddedExtension(extensionName)
	log.Debug("isEmbeddedExtension:" + extensionName + " =>" + strconv.FormatBool(isEmbeddedExtension))
	if err != nil {
		return err
	}
	var extensionPath string
	backupTaken, err := backupExtension(extensionName)
	if err != nil {
		return err
	}
	var errInstall error
	if isEmbeddedExtension {
		if zipPath != "" {
			fileInfo, _ := os.Stat(zipPath)
			if fileInfo.Size() != 0 {
				return errors.New("Extension name is already used by embedded extension")
			}
		}
		errInstall = CopyExtensionToEmbeddedExtensionPath(extensionName)
		extensionPath = filepath.Join(GetExtensionPathEmbedded(), extensionName)
	} else {
		if zipPath != "" {
			errInstall = Unzip(zipPath, GetExtensionPathCustom(), extensionName)
			if errInstall == nil {
				os.Remove(zipPath)
			}
		} else {
			errInstall = errors.New("the zipPath parameter is missing")
		}
		extensionPath = filepath.Join(GetExtensionPathCustom(), extensionName)
	}
	var errGen error
	if errInstall == nil {
		errGen = generateStatesFile(extensionName, extensionPath)
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
		return errors.New("Error Generate States file:" + errGen.Error() + "\nRegistration rolled back")
	}
	if errInstall != nil {
		return errors.New("Error Install:" + errInstall.Error() + "\nRegistration rolled back")
	}
	return nil
}

func generateStatesFile(extensionName, extensionPath string) error {
	log.Debug("Entering in... generateStatesFile")
	manifest, err := ioutil.ReadFile(filepath.Join(extensionPath, "extension-manifest.yml"))
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
	statesFileCfg, err := config.ParseYaml("states:")
	if err != nil {
		return err
	}
	err = statesFileCfg.Set("states", cfg.Root)
	if err != nil {
		return err
	}
	out, err := config.RenderYaml(statesFileCfg.Root)
	if err != nil {
		return err
	}
	log.Debug("states file content:\n" + out)
	return ioutil.WriteFile(filepath.Join(extensionPath, "statesFile-"+extensionName+".yml"), []byte(out), 0644)

}

//UnregisterExtension delete and extension, deletion of Embedded extension is not permitted.
func UnregisterExtension(extensionName string) error {
	log.Debug("Entering in... UnregisterExtension")
	log.Debug("extensionName:", extensionName)
	isEmbeddedExtension, err := IsEmbeddedExtension(extensionName)
	log.Debug("isEmbeddedExtension:" + strconv.FormatBool(isEmbeddedExtension))
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if isEmbeddedExtension {
		return errors.New("Deletion of embedded extension is not permitted")
	}
	log.Debug("IsCustomExtensionRegistered:" + strconv.FormatBool(IsCustomExtensionRegistered(extensionName)))
	if !IsCustomExtensionRegistered(extensionName) {
		return errors.New("This extension is not registered")
	}
	err = os.RemoveAll(filepath.Join(GetExtensionPathCustom(), extensionName))
	if err != nil {
		return err
	}
	return nil
}
