/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
*/
package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
	"github.com/IBM/commands-runner/api/i18n/i18nUtils"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	log.SetLevel(log.FatalLevel)
	var extensionPath string
	var langs string

	verifyLocalization := func(c *cli.Context) error {
		err := verifyLocalization(extensionPath, langs)
		if err != nil {
			log.Error(err.Error())
		}
		return err
	}

	app := cli.NewApp()
	//Overwrite some app parameters
	app.Usage = "client ..."
	app.Version = "1.0.0"
	app.Description = "Sample client"

	//Enrich with extra client commands
	app.Commands = []cli.Command{
		{
			Name:   "verify",
			Action: verifyLocalization,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "extension-path, p",
					Usage:       "The path to the extension, default '.' ",
					Destination: &extensionPath,
				},
				cli.StringFlag{
					Name:        "languages, l",
					Usage:       "CSV list of the languages to verify, default all available language will be tested",
					Destination: &langs,
				},
			},
		},
	}

	//Run the command
	errRun := app.Run(os.Args)
	if errRun != nil {
		fmt.Println(errRun.Error())
		os.Exit(1)
	}
}

func verifyLocalization(extensionPath string, languages string) error {
	var requestedLanguages []string
	if languages != "" {
		r := csv.NewReader(strings.NewReader(languages))
		records, err := r.ReadAll()
		if err != nil {
			return err
		}
		requestedLanguages = records[0]
	}
	err := i18nUtils.LoadTranslationFilesFromDir(filepath.Join(extensionPath, i18nUtils.I18nDirectory))
	if err != nil {
		return err
	}
	langs, err := i18nUtils.GetAllLanguageTags()
	if err != nil {
		return err
	}
	for _, lang := range requestedLanguages {
		if !i18nUtils.IsSupportedLanguage(lang) {
			return errors.New(lang + " is not supported! Please add it in the " + i18nUtils.I18nDirectory + " directory")
		}
	}
	log.Debugf("Langs: %v\n", langs)
	foundError := false
	for _, lang := range langs {
		fmt.Printf("Check lang '%s' ... ", lang.String())
		if requestedLanguages == nil || (requestedLanguages != nil && hasElem(requestedLanguages, lang.String())) {
			_, messagesNotFound, err := state.GetUIMetadataTranslated(extensionPath, []string{lang.String()})
			if err != nil {
				return errors.New("Error while translating to " + lang.String() + ":" + err.Error())
			}
			if len(messagesNotFound) > 0 {
				fmt.Printf("NOK\n")
				foundError = true
				for _, message := range messagesNotFound {
					fmt.Printf("Message not found - lang (%s) : %s\n", lang.String(), message)
				}
			} else {
				fmt.Printf("OK\n")
			}
		}
	}
	if foundError {
		return errors.New("The above messages were not found")
	}
	return nil
}

func hasElem(a []string, elem string) bool {
	for _, s := range a {
		if s == elem {
			return true
		}
	}
	return false
}
