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
package commandsRunner

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	oconfig "github.com/olebedev/config"
	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/properties"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/status"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/i18n/i18nUtils"
	cli "gopkg.in/urfave/cli.v1"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Force, RunningToFailed")
}

func validateToken(configDir string, protectedHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		setupResponse(&w, req)
		if (*req).Method == "OPTIONS" {
			return
		}

		//Retreive Authentication header from request
		Auth := req.Header.Get("Authorization")
		if Auth == "" {
			logger.AddCallerField().Error("Auth header not found")
			http.Error(w, "Token not provided", http.StatusForbidden)
			return
		}
		//Split to find the provided token
		receivedTokens := strings.Split(Auth, ":")
		receivedToken := receivedTokens[1]

		//Read official token
		token, err := ioutil.ReadFile(filepath.Join(configDir, global.TokenFileName))
		if err != nil {
			http.Error(w, "Token file not found:"+filepath.Join(configDir, global.TokenFileName), http.StatusNotFound)
			return
		}

		//Convert and trim token
		tokenS := string(token)
		tokenS = strings.TrimSuffix(tokenS, "\n")

		//Check if correct token
		if receivedToken != tokenS {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		//Forward to the handler
		protectedHandler.ServeHTTP(w, req)

	})
}

func AddHandler(pattern string, handler http.HandlerFunc, requireAuth bool) {
	log.Debug("Entering... AddHandler")
	log.Debug("pattern:" + pattern)
	log.Debug("serverConfigDir:" + global.ServerConfigDir)
	log.Debug("requireAuth:" + strconv.FormatBool(requireAuth))
	if requireAuth {
		http.HandleFunc(pattern, validateToken(global.ServerConfigDir, handler))
	} else {
		http.HandleFunc(pattern, handler)
	}
}

func readCommandsRunnerConfig(configDir string) error {
	log.Debug("Entering in... readConfig")
	raw, e := ioutil.ReadFile(filepath.Join(configDir, global.CommandsRunnerConfigFileName))
	if e == nil {
		uiConfigCfg, err := oconfig.ParseYamlBytes(raw)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		var properties properties.Properties
		properties, err = uiConfigCfg.Map("")
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		if val, ok := properties["port"]; ok {
			global.ServerPort = val.(string)
		}
		if val, ok := properties["port_ssl"]; ok {
			global.ServerPortSSL = val.(string)
		}
		if val, ok := properties["default_extension_name"]; ok {
			commandsRunner.SetDefaultExtensionName(val.(string))
		}
		//Move in the i18N file
		// if val, ok := properties["default_deployment_name"]; ok {
		// 	commandsRunner.SetDeploymentName(val.(string))
		// }
		if val, ok := properties["extension_embedded_file"]; ok {
			state.SetExtensionsEmbeddedFile(val.(string))
		}
		if val, ok := properties["embedded_extensions_repository_path"]; ok {
			err := state.SetEmbeddedExtensionsRepositoryPath(val.(string))
			if err != nil {
				log.Fatal(err)
			}
		}
		if val, ok := properties["extensions_path"]; ok {
			state.SetExtensionsPath(val.(string))
		}
		if val, ok := properties["extensions_logs_path"]; ok {
			state.SetExtensionsLogsPath(val.(string))
		}
		if val, ok := properties["about_url"]; ok {
			commandsRunner.SetAboutURL(val.(string))
		}
		if val, ok := properties["about"]; ok {
			commandsRunner.SetAbout(val.(string))
		}
		return nil
	}
	log.Info("No CommandsRunner config file found")
	return nil
}

func start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Fatal(r)
				logger.LogFile.Close()
			}
		}()
		log.Info("http://localhost:" + global.ServerPort)
		if err := http.ListenAndServe(":"+global.ServerPort, nil); err != nil {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()
	_, errCertPath := os.Stat(global.ServerCertificatePath)
	_, errKeyPath := os.Stat(global.ServerKeyPath)
	if errCertPath == nil && errKeyPath == nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Fatal(r)
					logger.LogFile.Close()
				}
			}()
			log.Info("https://localhost:" + global.ServerPortSSL)
			if err := http.ListenAndServeTLS(":"+global.ServerPortSSL, global.ServerCertificatePath, global.ServerKeyPath, nil); err != nil {
				log.Errorf("ListenAndServeTLS error: %v", err)
			}
		}()
	} else {
		log.Info("SSL not enabled as " + global.ServerCertificatePath + " or " + global.ServerKeyPath + " is not present.")
	}
	status.SetStatus(status.CMStatus, "Up")
}

func blockForever() {
	select {}
}

type InitFunc func(port string, portSSL string, configDir string, certificatePath string, keyPath string)

type PostStartFunc func(configDir string)

func Init(port string, portSSL string, configDir string, certificatePath string, keyPath string) {
	global.ServerPort = port
	global.ServerPortSSL = portSSL
	global.ServerConfigDir = configDir
	global.ServerCertificatePath = certificatePath
	global.ServerKeyPath = keyPath
	config.SetConfigPath(configDir)
	AddHandler("/cr/v1/state", state.HandleState, true)
	AddHandler("/cr/v1/state/", state.HandleState, true)
	AddHandler("/cr/v1/states", state.HandleStates, true)
	AddHandler("/cr/v1/engine", state.HandleEngine, true)
	AddHandler("/cr/v1/cr/", commandsRunner.HandleCR, true)
	AddHandler("/cr/v1/status", status.HandleStatus, true)
	AddHandler("/cr/v1/extension", state.HandleExtension, true)
	AddHandler("/cr/v1/extensions", state.HandleExtensions, true)
	AddHandler("/cr/v1/extensions/", state.HandleExtensions, true)
	AddHandler("/cr/v1/uimetadata", state.HandleUIMetadata, true)
	AddHandler("/cr/v1/uimetadatas", state.HandleUIMetadatas, true)
	AddHandler("/cr/v1/config", config.HandleConfig, true)
	AddHandler("/cr/v1/config/", config.HandleConfig, true)
	AddHandler("/cr/v1/template", state.HandleTemplate, true)
}

func ServerStart(preInit InitFunc, postInit InitFunc, preStart InitFunc, postStart PostStartFunc) {
	var configDir string
	var port string
	var portSSL string
	//	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//If Panic close the current log.
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
			logger.LogFile.Close()
		}
	}()

	app := cli.NewApp()
	app.Usage = "Commands Runner for installation"
	app.Description = "CLI to manage initial Commands Runner installation"

	app.Commands = []cli.Command{
		{
			Name:  "listen",
			Usage: "Launch the Config Manager server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "configDir, c",
					Usage:       "Config Directory",
					Destination: &configDir,
				},
				cli.StringFlag{
					Name:        "port, p",
					Usage:       "Port",
					Value:       global.DefaultPort,
					Destination: &port,
				},
				cli.StringFlag{
					Name:        "portSSL, pssl",
					Usage:       "PortSSL",
					Value:       global.DefaultPortSSL,
					Destination: &portSSL,
				},
			},
			Action: func(c *cli.Context) error {
				logMaxBackups := os.Getenv("CR_LOG_MAX_BACKUPS")
				maxBackups := 10
				var errMaxBackups error
				if logMaxBackups != "" {
					maxBackups, errMaxBackups = strconv.Atoi(logMaxBackups)
					if errMaxBackups != nil {
						maxBackups = 10
					}
				}

				logger.InitLogFile(configDir, maxBackups)

				out := io.MultiWriter(logger.LogFile, os.Stderr)
				log.SetOutput(out)
				if errMaxBackups != nil {
					log.Errorf("The provided maximum backups for log can not be converted to integer %s (default value used)", errMaxBackups.Error())
				}
				err := logger.LogFile.Rotate()
				if err != nil {
					log.Error(err.Error())
				}
				logLevel := os.Getenv("CR_TRACE")
				log.Printf("CR_TRACE: %s", logLevel)
				if logLevel == "" {
					logLevel = logger.DefaultLogLevel
				}
				level, err := log.ParseLevel(logLevel)
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				log.SetLevel(level)
				//Restore I18N files
				log.Info("Install i18n files")
				err = i18nUtils.RestoreFiles()
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				err = i18nUtils.LoadMessageFiles()
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				log.Info("Starting cm server")
				if configDir == "" {
					logger.AddCallerField().Error("Missing option -c to specif the directory where the config must be stored")
					return errors.New("Missing option -c to specif the directory where the config must be stored")
				}
				//check if path absolute
				if !filepath.IsAbs(configDir) {
					logger.AddCallerField().Fatal("The path of config must be absolute: " + configDir)
				}
				err = readCommandsRunnerConfig(configDir)
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				if preInit != nil {
					preInit(port, portSSL, configDir, filepath.Join(configDir, global.SSLCertFileName), filepath.Join(configDir, global.SSLKeyFileName))
				}
				Init(port, portSSL, configDir, filepath.Join(configDir, global.SSLCertFileName), filepath.Join(configDir, global.SSLKeyFileName))
				if postInit != nil {
					postInit(port, portSSL, configDir, filepath.Join(configDir, global.SSLCertFileName), filepath.Join(configDir, global.SSLKeyFileName))
				}
				err = state.DeleteFormerEmbeddedExtensions()
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				err = state.RegisterEmbededExtensions(true, true)
				if err != nil {
					logger.AddCallerField().Fatal(err.Error())
				}
				if preStart != nil {
					preStart(port, portSSL, configDir, filepath.Join(configDir, global.SSLCertFileName), filepath.Join(configDir, global.SSLKeyFileName))
				}
				start()
				if postStart != nil {
					postStart(configDir)
				}
				blockForever()
				return nil
			},
		},
	}
	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
