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
package commandsRunner

import (
	"archive/zip"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
)

func cleanup() {
	_ = os.RemoveAll("../test/resource/tmp")
	_ = os.Remove("../test/resource/tmp")
}

func assert(expected, actual string, t *testing.T) {
	if actual != expected {
		t.Errorf("expected \n%v actual \n%v", expected, actual)
	}
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func createFileUploadRequest(pathToFile, extensionName string, t *testing.T) *http.Request {
	var req *http.Request
	if pathToFile != "" {
		zipit("../test/resource/extensions/custom-extension", pathToFile)
		body, _ := os.Open(pathToFile)
		writer := multipart.NewWriter(body)
		req, _ = http.NewRequest("POST", "/cr/v1/extension/action=register", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Content-Disposition", "upload; filename="+filepath.Base(pathToFile))
	} else {
		req, _ = http.NewRequest("POST", "/cr/v1/extension/action=register", nil)
	}
	req.Header.Set("Extension-Name", extensionName)
	return req
}

func TestRegisterExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterExistingExtension")
	// Setup unit test file structure
	extensionManager.SetExtensionPath("../test/resource/tmp/")

	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../test/data/extensions/")
	extensionName := "dummy-extension"
	filename := "dummy-extension.zip"
	if _, err := os.Stat(extensionManager.GetExtensionPath()); os.IsNotExist(err) {
		err := os.Mkdir(extensionManager.GetExtensionPath(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(extensionManager.GetExtensionPathCustom()); os.IsNotExist(err) {
		err = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(filepath.Join(extensionManager.GetExtensionPathCustom(), extensionName)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(extensionManager.GetExtensionPathCustom(), extensionName), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	fileCreated, err := os.Create(filepath.Join(extensionManager.GetExtensionPathCustom(), filename))
	if err != nil {
		t.Fatal(err)
	}

	fileCreated.Close()

	// Create and handle request for unit test
	req := createFileUploadRequest("../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension "+extensionName+" already registered\n", rr.Body.String(), t)

	cleanup()
}

func TestRegisterNonExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterNonExistingExtension")

	//Setup filesystem
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	filename := "dummy-extension.zip"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("../test/resource/"+filename, "dummy-extension", t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	if _, err := os.Stat(filepath.Join(extensionManager.GetExtensionPathCustom(), "dummy-extension")); os.IsNotExist(err) {
		t.Errorf("project was not unzipped %v\n", err)
	}

	cleanup()
}

func TestRegisterCustomExtension(t *testing.T) {
	t.Log("Entering........... TestExtensionUnzip")
	// Dummy extensionManager.GetExtensionPath()
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "blahblahblah"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)

	req := createFileUploadRequest("../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(extensionManager.GetExtensionPathCustom(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "extension-manifest.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "/scripts/success.sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	//	cleanup()
}

func TestRegisterCustomExtensionWithIBMExtensionName(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering........... TestRegisterCustomExtensionWithIBMExtensionName")
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../test/resource/tmp/")

	//Setup filesystem
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension name is already used by "+extensionManager.EmbeddedExtensions+" extension\n", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtension")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionManager.SetEmbeddedExtensionsRepositoryPath("../test/repo_local/")

	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("", extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtensionFilesExists(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtensionFilesExists")
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionManager.SetEmbeddedExtensionsRepositoryPath("../test/repo_local/")
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathEmbedded(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("", extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(extensionManager.GetExtensionPathEmbedded(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}
	cleanup()
}

func TestDeletionEndpointExists(t *testing.T) {
	t.Log("Entering........... TestExtensionDeletion")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionName := "dummy-extension"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(filepath.Join(extensionManager.GetExtensionPathCustom(), extensionName), 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("Delete returned: %v", rr.Code)
	}

	cleanup()
}

func TestDeletionExtensionExists(t *testing.T) {
	t.Log("Entering........... TestDeletionExtensionExists")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionName := "dummy-extension2"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom()+"/dummy-extension", 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 500 {
		t.Fatalf("Extension should not exists. Status code: %v", rr.Code)
	}

	cleanup()
}

func TestDeletionFromFileSystem(t *testing.T) {
	t.Log("Entering........... TestDeletionFromFileSystem")
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	extensionName := "dummy-extension"
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	err := os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(extensionManager.GetExtensionPathCustom()+"/dummy-extension", 0777)
	if err != nil {
		t.Log(err)
	}
	os.Create(extensionManager.GetExtensionPathCustom() + dontDeleteFile)
	os.Create(extensionManager.GetExtensionPathCustom() + "/dummy-extension/" + deleteFile)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	file, err := os.Stat(extensionManager.GetExtensionPathCustom() + extensionName)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(extensionManager.GetExtensionPathCustom() + deleteFile)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(extensionManager.GetExtensionPathCustom() + dontDeleteFile)
	if file == nil {
		t.Errorf("The extension, %s, was not suppose to be deleted", extensionName)
	}

	cleanup()
}

func TestListEndpointExists(t *testing.T) {
	req, err := http.NewRequest("GET", "/cr/v1/extensions/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}
	cleanup()
}

func setupFileStructureLists() {
	extensionManager.SetExtensionEmbeddedFile("../test/resource/extensions/test-extensions.txt")
	extensions := [4]string{"dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"}
	extensionsIBM := [4]string{"IBM-extension1", "IBM-extension2"}
	extensionManager.SetExtensionPath("../test/resource/tmp/")
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	_ = os.Mkdir(extensionManager.GetExtensionPath(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(extensionManager.GetExtensionPathEmbedded(), 0777)
	for _, extension := range extensions {
		_ = os.Mkdir(filepath.Join(extensionManager.GetExtensionPathCustom(), extension), 0777)
	}
	for _, extension := range extensionsIBM {
		_ = os.Mkdir(filepath.Join(extensionManager.GetExtensionPathEmbedded(), extension), 0777)
	}
	os.Create(filepath.Join(extensionManager.GetExtensionPathCustom(), dontDeleteFile))
	os.Create(filepath.Join(extensionManager.GetExtensionPathCustom(), deleteFile))
	os.Mkdir(filepath.Join(extensionManager.GetExtensionPathCustom(), extensions[0], "/do-not-return-embedded-dir"), 0777)
}

func TestListAllExensions(t *testing.T) {
	t.Log("TESTING..................... TestListAllExensions")
	setupFileStructureLists()

	req, err := http.NewRequest("GET", "/cr/v1/extensions?catalog=false", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}

	var extensions extensionManager.Extensions
	extensions.Extensions = make(map[string]extensionManager.Extension)
	extension1 := &extensionManager.Extension{
		Name: "dummy-extension1",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension1.Name] = *extension1
	extension2 := &extensionManager.Extension{
		Name: "dummy-extension2",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension2.Name] = *extension2
	extension3 := &extensionManager.Extension{
		Name: "dummy-extension3",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension3.Name] = *extension3
	extension4 := &extensionManager.Extension{
		Name: "dummy-extension4",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension4.Name] = *extension4
	extension5 := &extensionManager.Extension{
		Name: "IBM-extension1",
		Type: extensionManager.EmbeddedExtensions,
	}
	extensions.Extensions[extension5.Name] = *extension5
	extension6 := &extensionManager.Extension{
		Name: "IBM-extension2",
		Type: extensionManager.EmbeddedExtensions,
	}
	extensions.Extensions[extension6.Name] = *extension6
	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"],"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListCustomerExensionsWithEmbeddedFolders(t *testing.T) {
	t.Log("TESTING..................... TestListCustomerExensionsWithEmbeddedFolders")
	log.SetLevel(log.DebugLevel)
	setupFileStructureLists()

	req, err := http.NewRequest("GET", "/cr/v1/extensions?filter="+extensionManager.CustomExtensions, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}

	var extensions extensionManager.Extensions
	extensions.Extensions = make(map[string]extensionManager.Extension)
	extension1 := &extensionManager.Extension{
		Name: "dummy-extension1",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension1.Name] = *extension1
	extension2 := &extensionManager.Extension{
		Name: "dummy-extension2",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension2.Name] = *extension2
	extension3 := &extensionManager.Extension{
		Name: "dummy-extension3",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension3.Name] = *extension3
	extension4 := &extensionManager.Extension{
		Name: "dummy-extension4",
		Type: extensionManager.CustomExtensions,
	}
	extensions.Extensions[extension4.Name] = *extension4
	expected, _ := json.MarshalIndent(&extensions, "", "  ")

	//	expected := `{"extensions":{"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListIBMExensions(t *testing.T) {
	t.Log("TESTING..................... TestListIBMExensions")
	setupFileStructureLists()
	extensionManager.SetExtensionPath("../test/resource/tmp/")

	req, err := http.NewRequest("GET", "/cr/v1/extensions?filter="+extensionManager.EmbeddedExtensions+"&catalog=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}

	var extensions extensionManager.Extensions
	extensions.Extensions = make(map[string]extensionManager.Extension)
	extension1 := &extensionManager.Extension{
		Name: "cfp-ext-template",
		Type: extensionManager.EmbeddedExtensions,
	}
	extension2 := &extensionManager.Extension{
		Name: "cfp-ext-template-auto-location",
		Type: extensionManager.EmbeddedExtensions,
	}

	extensions.Extensions[extension1.Name] = *extension1
	extensions.Extensions[extension2.Name] = *extension2
	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"]}}`
	log.Debug(expected)
	log.Debug([]byte(rr.Body.String()))
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	cleanup()
}
