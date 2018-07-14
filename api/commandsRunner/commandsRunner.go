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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extension"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/status"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/uiConfig"
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

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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
			log.Info(receivedToken)
			log.Info(tokenS)
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

func start() {
	go func() {
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
			log.Errorf("ListenAndServe error: %v", err)
		}
	}()
	go func() {
		if err := http.ListenAndServeTLS(":"+serverPortSSL, serverCertificatePath, serverKeyPath, nil); err != nil {
			log.Errorf("ListenAndServeTLS error: %v", err)
		}
	}()
	status.SetStatus(status.CMStatus, "Up")
	// if _, err := os.Stat(serverCertificatePath); err == nil {
	// 	if _, err := os.Stat(serverKeyPath); err == nil {

	// 		log.Fatal(http.ListenAndServeTLS(":"+serverPortSSL, serverCertificatePath, serverKeyPath, nil))
	// 	}
	// }
	blockForever()
}

func blockForever() {
	select {}
}

type InitFunc func(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string)

func Init(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string) {
	serverPort = port
	serverPortSSL = portSSL
	serverConfigDir = configDir
	serverCertificatePath = certificatePath
	serverKeyPath = keyPath
	config.SetConfigPath(configDir)
	state.SetStatePath(stateFilePath)
	AddHandler("/cr/v1/state", state.HandleState, true)
	AddHandler("/cr/v1/state/", state.HandleState, true)
	AddHandler("/cr/v1/states", state.HandleStates, true)
	AddHandler("/cr/v1/engine", state.HandleEngine, true)
	AddHandler("/cr/v1/pcm/", commandsRunner.HandleCR, true)
	AddHandler("/cr/v1/status", status.HandleStatus, true)
	AddHandler("/cr/v1/extension", extension.HandleExtension, true)
	AddHandler("/cr/v1/extensions", extension.HandleExtensions, true)
	AddHandler("/cr/v1/extensions/", extension.HandleExtensions, true)
	AddHandler("/cr/v1/uiconfig/", uiConfig.HandleUIConfig, true)
	AddHandler("/cr/v1/config", config.HandleConfig, true)
	AddHandler("/cr/v1/config/", config.HandleConfig, true)
}

func ServerStart(preInitFunc InitFunc, postInitFunc InitFunc, preStartFunc InitFunc) {
	var configDir string
	var port string
	var portSSL string
	var statesPath string

	//	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
					Name:        "statePath, s",
					Usage:       "Path of the state file",
					Destination: &statesPath,
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
				commandsRunner.LogPath = configDir + string(filepath.Separator) + "commands-runner.log"
				file, _ := os.Create(commandsRunner.LogPath)
				out := io.MultiWriter(file, os.Stderr)
				log.SetOutput(out)
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
					preInitFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, statesPath)
				}
				Init(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, statesPath)
				if postInitFunc != nil {
					postInitFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, statesPath)
				}
				if preStartFunc != nil {
					preStartFunc(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName, statesPath)
				}
				start()
				return nil
			},
		},
	}
	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(1)
	}

}
