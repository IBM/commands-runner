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

package state

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

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
	Type      string    `yaml:"type" json:"type"`
	Version   string    `yaml:"version" json:"version"`
	CallState CallState `yaml:"call_state" json:"call_state"`
}

type CallState struct {
	StatesToRerun  []string `yaml:"states_to_rerun" json:"states_to_rerun"`
	PreviousStates []string `yaml:"previous_states" json:"previous_states"`
	NextStates     []string `yaml:"next_states" json:"next_states"`
}

/*
Extensions is a map of Extension
*/
type Extensions struct {
	Extensions map[string]Extension `yaml:"extensions" json:"extensions"`
}

/*
Init initialise the extensionManager package.
embeddedExtensionDescriptor the path to the file listing the embedded extension names.
embeddedExtensionsRepositoryPath the directory path where the extensions are located. This path will be extended with "embeded" or "custom" depending of the type of extension.
extensionPath the directory path where the extension will be copied by the registration process. This path will be extended with "embeded" or "custom" depending of the type of extension.
extensionLogsPath the directory path where the logs will be save.
This path is relative to extension deployment location which is <extensionPath>/<embedded|custom>/<extensionName>.
This path will be extended with "embeded" or "custom" depending of the type of extension.
*/
func InitExtensions(embeddedExtensionDescriptor string, embeddedExtensionsRepositoryPath string, extensionPath string, extensionLogsPath string) {
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
	extensionPath += GetRelativeExtensionPath(extensionName)
	return extensionPath
}

//IsExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsExtensionRegistered(extensionName string) bool {
	log.Debug("Entering in... IsExtensionRegistered")
	return IsEmbeddedExtensionRegistered(extensionName) || IsCustomExtensionRegistered(extensionName)
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
	extensions, err := ListEmbeddedExtensions()
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	log.Debug(extensions)
	extension, ok := extensions.Extensions[extensionName]
	if !ok {
		err := errors.New("Extension name: " + extensionName + " not found")
		log.Error(err.Error())
		return "", err
	}
	if extension.Version == "" {
		return filepath.Join(embeddedExtensionsRepositoryPath, extensionName), nil
	}
	return filepath.Join(embeddedExtensionsRepositoryPath, extensionName, extension.Version), nil
	// files, err := ioutil.ReadDir(filepath.Join(embeddedExtensionsRepositoryPath, extensionName))
	// if err != nil {
	// 	log.Debug(err.Error())
	// 	return "", err
	// }
	// if len(files) == 0 {
	// 	return "", errors.New("extension directory " + extensionName + " is empty")
	// }
	// if len(files) == 1 && files[0].IsDir() {
	// 	return filepath.Join(embeddedExtensionsRepositoryPath, extensionName, files[0].Name()), nil
	// }
	// return filepath.Join(embeddedExtensionsRepositoryPath, extensionName), nil
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
			var callState CallState
			extensionName := file.Name()
			if extensionPath == GetExtensionPathCustom() {
				extension.Type = CustomExtensions
			} else {
				extension.Type = EmbeddedExtensions
			}
			log.Debug("extension.Name: " + extensionName)
			manifestPath := filepath.Join(extensionPath, extensionName, "extension-manifest.yml")
			log.Debug("manifestPath: " + manifestPath)
			manifestBytes, err := ioutil.ReadFile(manifestPath)
			if err != nil {
				return &extensionList, err
			}
			cfg, err := config.ParseYaml(string(manifestBytes))
			if err != nil {
				return &extensionList, err
			}
			stateCfg, err := cfg.Get("call_state")
			if err == nil {
				stateString, err := config.RenderYaml(stateCfg.Root)
				if err != nil {
					return &extensionList, err
				}
				log.Debug("call_state: " + stateString)
				err = yaml.Unmarshal([]byte(stateString), &callState)
				if err != nil {
					return &extensionList, err
				}
				log.Debugf("%v", callState)
				extension.CallState = callState
			}
			extensionList.Extensions[extensionName] = extension
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
	//extensionList.Extensions = make(map[string]Extension)
	log.Debug("extensionEmbeddedFile:" + extensionEmbeddedFile)
	if extensionEmbeddedFile == "" {
		return &extensionList, nil
	}
	resource, err := ioutil.ReadFile(extensionEmbeddedFile)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	err = yaml.Unmarshal(resource, &extensionList)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	for key, extension := range extensionList.Extensions {
		log.Debug("key: " + key)
		extension.Type = EmbeddedExtensions
		extensionList.Extensions[key] = extension
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
	backupPath := filepath.Join("/tmp", extensionName) + string(filepath.Separator)
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
	backupPath := filepath.Join("/tmp", extensionName) + string(filepath.Separator)
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
	log.Debug("Entering in... registerEmbededExtensions")
	extensions, err := ListEmbeddedExtensions()
	if err != nil {
		return err
	}
	for key := range extensions.Extensions {
		log.Debug("key: " + key)
		err := RegisterExtension(key, "", force)
		if err != nil {
			return err
		}
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
	var errGenStatesFile error
	if errInstall == nil {
		errGenStatesFile = GenerateStatesFile(extensionName, extensionPath)
		//Failure to generate the template files must not stop the installation.
		GenerateTemplateFiles(extensionName, extensionPath)
	}
	if errInstall != nil || errGenStatesFile != nil {
		if backupTaken {
			restoreExtension(extensionName)
		} else {
			os.RemoveAll(extensionPath)
			os.Remove(extensionPath)
		}
	}
	if errGenStatesFile != nil {
		return errors.New("Error Generate States file:" + errGenStatesFile.Error() + "\nRegistration rolled back")
	}
	if errInstall != nil {
		return errors.New("Error Install:" + errInstall.Error() + "\nRegistration rolled back")
	}
	return nil
}

func GenerateStatesFile(extensionName string, extensionPath string) error {
	log.Debug("Entering in... GenerateStatesFile")
	log.Debug("Extension:" + extensionName)
	manifestPath := filepath.Join(extensionPath, "extension-manifest.yml")
	input, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var inYaml map[string]interface{}
	inYaml = make(map[string]interface{}, 0)
	err = yaml.Unmarshal(input, &inYaml)
	if err != nil {
		return err
	}
	statesUpdateMode := "merge"
	if val, ok := inYaml["states_update_mode"]; ok {
		statesUpdateMode = strings.ToLower(val.(string))
	}
	newStatesB, err := global.ExtractKey(filepath.Join(extensionPath, "extension-manifest.yml"),
		"states")
	if err != nil {
		return err
	}
	switch statesUpdateMode {
	case "merge":
		sm, err := GetStatesManager(extensionName)
		if err != nil {
			return err
		}
		var newStates States
		err = yaml.Unmarshal(newStatesB, &newStates)
		if err != nil {
			return err
		}
		err = sm.SetStates(newStates, false)
		if err != nil {
			return err
		}
	case "replace":
		err = ioutil.WriteFile(filepath.Join(extensionPath, global.StatesFileName), newStatesB, 0644)
		if err != nil {
			return err
		}
	case "new":
		ext := filepath.Ext(global.StatesFileName)
		pathWithoutExt := strings.TrimSuffix(global.StatesFileName, "."+ext)
		newStatesFileName := pathWithoutExt + "-new." + ext
		err = ioutil.WriteFile(filepath.Join(extensionPath, newStatesFileName), newStatesB, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateTemplateFiles(extensionName string, extensionPath string) error {
	log.Debug("Entering in... GenerateTemplateFiles")
	manifestPath := filepath.Join(extensionPath, "extension-manifest.yml")
	input, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var inYaml map[string]interface{}
	inYaml = make(map[string]interface{}, 0)
	err = yaml.Unmarshal(input, &inYaml)
	if err != nil {
		return err
	}
	var uiMetaData map[interface{}]interface{}
	if val, ok := inYaml["ui_metadata"]; ok {
		uiMetaData, ok = val.(map[interface{}]interface{})
		if !ok {
			return errors.New("ui_metadata is not a map[interface{}]interface{}")
		}
	}
	//loop on configurations
	for uiMetadataName := range uiMetaData {
		data, err := GenerateUIMetaDataTemplate(extensionName, uiMetadataName.(string))
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(extensionPath, global.ConfigRootKey+"_"+uiMetadataName.(string)+"_template.yml"), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
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
