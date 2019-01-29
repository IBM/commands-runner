/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package i18nUtils

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	yaml "gopkg.in/yaml.v2"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/i18n/i18nBinData"
)

const I18nDirectory = "i18n"

//Bundle i18n bundle holder
var Bundle *i18n.Bundle

//GetI18nDir returns the directory where the i18n files are.
func GetI18nDir() (string, error) {
	launchingDir, err := global.GetExecutableDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(launchingDir, I18nDirectory), nil
}

//RestoreFiles loaded i18nBinData by go-bindata into the  <directory where the code is running>/i18n
func RestoreFiles() error {
	log.Debug("Entering in.... RestoreFiles")
	i18nDir, err := GetI18nDir()
	if err != nil {
		return err
	}
	log.Info("i18n file directory: " + i18nDir)
	err = os.MkdirAll(i18nDir, 0700)
	if err != nil {
		return err
	}
	for _, assetName := range i18nBinData.AssetNames() {
		log.Debug("Asset Name:" + assetName)
		err = i18nBinData.RestoreAsset(i18nDir, assetName)
		if err != nil {
			return err
		}
	}
	return nil
}

//LoadMessageFiles load translation files located in the <directory where the code is running>/i18n
func LoadMessageFiles() error {
	i18nInternalDir, err := GetI18nDir()
	if err != nil {
		return err
	}
	err = LoadTranslationFilesFromDir(i18nInternalDir)
	if err != nil {
		return err
	}
	return nil
}

//Load into the Bundle with the translation file located in i18nDir
func LoadTranslationFilesFromDir(i18nDir string) error {
	//Create bundle holding all messages for i18n
	if Bundle == nil {
		Bundle = &i18n.Bundle{DefaultLanguage: language.English}
	}
	Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	Bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	files, err := filepath.Glob(filepath.Join(i18nDir, "*"))
	if err != nil {
		log.Error("Error searching for translation files", err.Error())
		return err
	}
	for _, file := range files {
		log.Debug("Load translation file:" + file)
		Bundle.MustLoadMessageFile(file)
	}
	return nil
}

func GetLangs(req *http.Request) []string {
	langs := make([]string, 0)
	lang := req.FormValue("lang")
	if lang != "" {
		langs = append(langs, lang)
	}
	accept := req.Header.Get("Accept-Language")
	if accept != "" {
		langs = append(langs, accept)
	}
	langs = append(langs, global.DefaultLanguage)
	return langs
}

func Translate(key string, defaultTranslation string, langs []string) (string, error) {
	if Bundle == nil {
		log.Debug("Load translation files")
		err := RestoreFiles()
		if err != nil {
			return "", err
		}
		err = LoadMessageFiles()
		if err != nil {
			return "", err
		}
	}
	localizer := i18n.NewLocalizer(Bundle, langs...)
	translation, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: defaultTranslation,
		},
	})
	return translation, err
}

func GetAllLanguageTags() ([]language.Tag, error) {
	if Bundle == nil {
		err := RestoreFiles()
		if err != nil {
			return nil, err
		}
		err = LoadMessageFiles()
		if err != nil {
			return nil, err
		}
	}
	return Bundle.LanguageTags(), nil
}

func IsSupportedLanguage(lang string) bool {
	tags, _ := GetAllLanguageTags()
	for _, tagAux := range tags {
		if tagAux.String() == lang {
			return true
		}
	}
	return false
}
