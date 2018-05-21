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
package cm

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"

	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/global"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/logger"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/resourceManager"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/statusManager"
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

func validateToken(bmxConfigDir string, protectedHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//Set debug level
		/*
			logLevel := os.Getenv("CM_TRACE")
			log.Printf("CM_TRACE: %s", logLevel)
			if logLevel == "" {
				logLevel = logger.DefaultLogLevel
			}
			level, err := log.ParseLevel(logLevel)
			if err != nil {
				log.Error(err.Error())
			} else {
				log.SetLevel(level)
			}
		*/
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
		token, err := ioutil.ReadFile(bmxConfigDir + "/" + global.TokenFileName)
		if err != nil {
			http.Error(w, "Token file not found:"+bmxConfigDir+"/"+global.TokenFileName, http.StatusNotFound)
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

func Init(port string, portSSL string, configDir string, certificatePath string, keyPath string, stateFilePath string) {
	serverPort = port
	serverPortSSL = portSSL
	serverConfigDir = configDir
	serverCertificatePath = certificatePath
	serverKeyPath = keyPath
	SetStatePath(stateFilePath)
	AddHandler("/cm/v1/state", handleState, true)
	AddHandler("/cm/v1/state/", handleState, true)
	AddHandler("/cm/v1/states", handleStates, true)
	AddHandler("/cm/v1/engine", handleEngine, true)
	AddHandler("/cm/v1/pcm/", handlePCM, true)
	AddHandler("/cm/v1/extension", handleExtension, true)
	AddHandler("/cm/v1/extensions/", handleExtensions, true)
}

func main() {
	var bmxConfigDir string
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
					Destination: &bmxConfigDir,
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
				if bmxConfigDir == "" {
					logger.AddCallerField().Error("Missing option -c to specif the directory where the config must be stored")
					return errors.New("Missing option -c to specif the directory where the config must be stored")
				}

				//check if path absolute
				if !filepath.IsAbs(bmxConfigDir) {
					log.Fatal("The path of bmxConfig must be absolute: " + bmxConfigDir)
				}

				Init(port, portSSL, bmxConfigDir, bmxConfigDir+"/"+global.SSLCertFileName, bmxConfigDir+"/"+global.SSLKeyFileName, pieStatesPath)

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
