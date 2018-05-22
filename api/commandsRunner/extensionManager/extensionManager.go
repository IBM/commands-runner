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
var extensionDirEmbedded = EmbeddedExtensions + "/"
var extensionPathEmbedded = extensionPath + extensionDirEmbedded
var extensionDirCustom = CustomExtensions + "/"
var extensionPathCustom = extensionPath + extensionDirCustom

var extensionLogsPath = "/data/logs/extensions/"
var extensionLogsDirEmbedded = extensionDirEmbedded
var extensionLogsPathEmbedded = extensionLogsPath + extensionLogsDirEmbedded
var extensionLogsDirCustom = "custom/"
var extensionLogsPathCustom = extensionLogsPath + extensionLogsDirCustom

type Extension struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Extensions struct {
	Extensions map[string]Extension `json:"extensions"`
}

func SetEmbeddedExtensionsRepositoryPath(_embeddedExtensionsRepositoryPath string) {
	embeddedExtensionsRepositoryPath = _embeddedExtensionsRepositoryPath
}

//Set the extensionPath
func SetExtensionPath(_extensionPath string) {
	extensionPath = _extensionPath
	extensionPathEmbedded = _extensionPath + extensionDirEmbedded
	extensionPathCustom = _extensionPath + extensionDirCustom
}

func GetExtensionPath() string {
	return extensionPath
}

func GetExtensionPathEmbedded() string {
	return extensionPathEmbedded
}

func GetExtensionPathCustom() string {
	return extensionPathCustom
}

func GetExtensionLogsPathEmbedded() string {
	return extensionLogsPathEmbedded
}

func GetExtensionLogsPathCustom() string {
	return extensionLogsPathCustom
}

func GetRepoLocalPath() string {
	return embeddedExtensionsRepositoryPath
}

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
		extensionPath = GetExtensionPathEmbedded() + extensionName + string(filepath.Separator)
	} else {
		extensionPath = GetExtensionPathCustom() + extensionName + string(filepath.Separator)
	}
	return extensionPath, nil
}

func GetRelativeExtensionPath(extensionName string) string {
	log.Debug("Entering in... GetRelativeExtensionPath")
	var extensionPath string
	isEmbeddedExtension, _ := IsEmbeddedExtension(extensionName)
	log.Debug("isEmbeddedExtension:" + extensionName + " =>" + strconv.FormatBool(isEmbeddedExtension))
	if isEmbeddedExtension {
		extensionPath = extensionDirEmbedded + extensionName + string(filepath.Separator)
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
	}
	return extensionPath
}

func SetExtensionEmbeddedFile(_extensionEmbeddedFile string) {
	log.Debug("Entering in... SetExtensionEmbeddedFile")
	extensionEmbeddedFile = _extensionEmbeddedFile
}

//IsExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsExtensionRegistered(extensionName string) bool {
	log.Debug("Entering in... IsExtensionRegistered")
	return extensionName == global.CloudFoundryPieName || IsEmbeddedExtensionRegistered(extensionName) || IsCustomExtensionRegistered(extensionName)
}

//IsCustomExtensionRegistered Check if an extension is register by browzing the extensions directory
func IsCustomExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsCustomExtensionRegistered")
	if _, err := os.Stat(GetExtensionPathCustom() + filename); os.IsNotExist(err) {
		return false
	}
	return true
}

//IsEmbeddedxtensionRegistered Check if an extension is register by browzing the extensions directory
func IsEmbeddedExtensionRegistered(filename string) bool {
	log.Debug("Entering in... IsEmbeddedExtensionRegistered")
	if _, err := os.Stat(GetExtensionPathEmbedded() + filename); os.IsNotExist(err) {
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

func getEmbeddedExtensionRepoPath(extensionName string) (string, error) {
	log.Debug("Entering in... getEmbeddedExtensionRepoPath")
	log.Debug("extensionName:" + extensionName)
	log.Debug("embeddedExtensionsRepositoryPath:" + embeddedExtensionsRepositoryPath)
	files, err := ioutil.ReadDir(embeddedExtensionsRepositoryPath + extensionName)
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("No version available for embedded extension " + extensionName)
	}
	return embeddedExtensionsRepositoryPath + extensionName + string(filepath.Separator) + files[0].Name() + string(filepath.Separator), nil
}

func CopyExtensionToEmbeddedExtensionPath(extensionName string) error {
	log.Debug("Entering in... CopyExtensionToEmbeddedExtensionPath")
	destDir := GetExtensionPathEmbedded() + extensionName
	extensionRepoPath, err := getEmbeddedExtensionRepoPath(extensionName)
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
		log.Debug("path:" + path)
		newPath := strings.Replace(path, extensionRepoPath, GetExtensionPathEmbedded()+extensionName, 1)
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

//IsEmbeddedExtension Check if extensionName is a Embedded extension
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
		return nil, err
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
		return nil, errors.New("extensionEmbeddedFile not defined")
	}
	//resource, err := ioutil.ReadFile(extensionEmbeddedFile)
	//resource, err := resourceManager.Asset(extensionEmbeddedFile)
	cfg, err := config.ParseYamlFile(extensionEmbeddedFile)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	extensionListFound, err := cfg.List("extensions")
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	for index, _ := range extensionListFound {
		extensionSpecification, err := cfg.String("extensions." + strconv.Itoa(index) + ".extension")
		if err != nil {
			log.Debug(err.Error())
			return nil, err
		}
		extensionFields := strings.Split(extensionSpecification, ":")
		var extension Extension
		extension.Name = extensionFields[2]
		extension.Type = EmbeddedExtensions
		extensionList.Extensions[extension.Name] = extension

	}
	/*
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
	*/
	return &extensionList, nil
}

//ListExtensions list all extensions
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
		extensionPath = GetExtensionPathEmbedded() + extensionName
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
	err = os.RemoveAll(GetExtensionPathCustom() + extensionName)
	if err != nil {
		return err
	}
	return nil
}
