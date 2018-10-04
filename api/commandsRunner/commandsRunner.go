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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/commandsRunner"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/config"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/logger"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/state"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/status"
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
	log.Debug("requireAuth:" + strconv.FormatBool(requireAuth))
	if requireAuth {
		http.HandleFunc(pattern, validateToken(serverConfigDir, handler))
	} else {
		http.HandleFunc(pattern, handler)
	}
}

func start() {
	go func() {
		log.Info("http://localhost:" + serverPort)
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
			log.Errorf("ListenAndServe error: %v", err)
		}
	}()
	_, errCertPath := os.Stat(serverCertificatePath)
	_, errKeyPath := os.Stat(serverKeyPath)
	if errCertPath == nil && errKeyPath == nil {
		log.Info("https://localhost:" + serverPortSSL)
		go func() {
			if err := http.ListenAndServeTLS(":"+serverPortSSL, serverCertificatePath, serverKeyPath, nil); err != nil {
				log.Errorf("ListenAndServeTLS error: %v", err)
			}
		}()
	} else {
		log.Info("SSL not enabled as " + serverCertificatePath + " or " + serverKeyPath + " is not present.")
	}
	status.SetStatus(status.CMStatus, "Up")
}

func blockForever() {
	select {}
}

type InitFunc func(port string, portSSL string, configDir string, certificatePath string, keyPath string)

type PostStartFunc func(configDir string)

func Init(port string, portSSL string, configDir string, certificatePath string, keyPath string) {
	serverPort = port
	serverPortSSL = portSSL
	serverConfigDir = configDir
	serverCertificatePath = certificatePath
	serverKeyPath = keyPath
	config.SetConfigPath(configDir)
	AddHandler("/cr/v1/state", state.HandleState, true)
	AddHandler("/cr/v1/state/", state.HandleState, true)
	AddHandler("/cr/v1/states", state.HandleStates, true)
	AddHandler("/cr/v1/engine", state.HandleEngine, true)
	AddHandler("/cr/v1/pcm/", commandsRunner.HandleCR, true)
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
				commandsRunner.LogPath = configDir + string(filepath.Separator) + "commands-runner.log"
				file, _ := os.Create(commandsRunner.LogPath)
				out := io.MultiWriter(file, os.Stderr)
				log.SetOutput(out)
				logLevel := os.Getenv("CR_TRACE")
				log.Printf("CR_TRACE: %s", logLevel)
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
				if preInit != nil {
					preInit(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName)
				}
				Init(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName)
				if postInit != nil {
					postInit(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName)
				}
				if preStart != nil {
					preStart(port, portSSL, configDir, configDir+"/"+global.SSLCertFileName, configDir+"/"+global.SSLKeyFileName)
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
