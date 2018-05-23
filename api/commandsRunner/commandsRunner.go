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
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/configManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/resourceManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/statusManager"
)

const COPYRIGHT string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

var serverConfigDir string
var serverPort string
var serverPortSSL string
var serverCertificatePath string
var serverKeyPath string

func validateToken(configDir string, protectedHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		token, err := ioutil.ReadFile(configDir + "/" + global.TokenFileName)
		if err != nil {
			http.Error(w, "Token file not found:"+configDir+"/"+global.TokenFileName, http.StatusNotFound)
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
	log.Debug("serverConfigDir:" + serverConfigDir)
	if requireAuth {
		http.HandleFunc(pattern, validateToken(serverConfigDir, handler))
	} else {
		http.HandleFunc(pattern, handler)
	}
}

func Start() {
	go func() {
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()
	statusManager.SetStatus(statusManager.CMStatus, "Up")
	log.Fatal(http.ListenAndServeTLS(":"+serverPortSSL, serverCertificatePath, serverKeyPath, nil))
}

type InitFunc func(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string)

func Init(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string) {
	serverPort = port
	serverPortSSL = portSSL
	serverConfigDir = configDir
	serverCertificatePath = certificatePath
	serverKeyPath = keyPath
	configManager.SetConfigPath(configDir)
	SetStatePath(stateFilePath)
	AddHandler("/cr/v1/state", handleState, true)
	AddHandler("/cr/v1/state/", handleState, true)
	AddHandler("/cr/v1/states", handleStates, true)
	AddHandler("/cr/v1/engine", handleEngine, true)
	AddHandler("/cr/v1/pcm/", handlePCM, true)
	AddHandler("/cr/v1/status", handleStatus, true)
	AddHandler("/cr/v1/extension", handleExtension, true)
	AddHandler("/cr/v1/extensions/", handleExtensions, true)
	AddHandler("/cr/v1/uiconfig/", handleUIConfig, true)
	AddHandler("/cr/v1/config/", handleConfig, true)
}

func NewApp(preInitFunc InitFunc, postInitFunc InitFunc, preStartFunc InitFunc) {
	var configDir string
	var port string
	var portSSL string
	var pieStatesPath string

	//	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := cli.NewApp()
	app.Usage = "Config Manager for Cloud Foundry installation"
	raw, e := resourceManager.Asset("VERSION")
	if e != nil {
		log.Fatal("Version not found")
	}
	app.Version = string(raw)
	app.Description = "CLI to manage initial Cloud Foundry installation"

	app.Commands = []cli.Command{
		{
			Name:   "listen",
			Hidden: true,
			Usage:  "Launch the Config Manager server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "configDir, c",
					Usage:       "Config Directory",
					Destination: &configDir,
				},
				cli.StringFlag{
					Name:        "statePath, s",
					Usage:       "Path of the state file",
					Destination: &pieStatesPath,
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
				logLevel := os.Getenv("CM_TRACE")
				log.Printf("CM_TRACE: %s", logLevel)
				if logLevel == "" {
					logLevel = logger.DefaultLogLevel
				}
				level, err := log.ParseLevel(logLevel)
				if err != nil {
					log.Fatal(err.Error())
				}
				log.SetLevel(level)
				log.Info("Starting cm server")
				if configDir == "" {
					logger.AddCallerField().Error("Missing option -c to specif the directory where the config must be stored")
					return errors.New("Missing option -c to specif the directory where the config must be stored")
				}

				//check if path absolute
				if !filepath.IsAbs(configDir) {
					log.Fatal("The path of config must be absolute: " + configDir)
				}

				if preInitFunc != nil {
					preInitFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, pieStatesPath)
				}
				Init(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, pieStatesPath)
				if postInitFunc != nil {
					postInitFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, pieStatesPath)
				}
				if preStartFunc != nil {
					preStartFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, pieStatesPath)
				}
				Start()
				return nil
			},
		},
	}

	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
