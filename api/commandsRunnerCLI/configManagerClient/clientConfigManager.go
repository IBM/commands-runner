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
package configManagerClient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

const copyright string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

const configFile = ".commandsRunner.conf"

var configFilePath string

var homeDir string

type ConfigManagerClient struct {
	URL          string `json:"url"`
	OutputFormat string `json:"output_format"`
	Timeout      int    `json:"timeout"`
	CACertPath   string `json:"ca_cert_path"`
	Token        string `json:"token"`
	InsecureSSL  bool
	rootCertPEM  []byte
	protocol     string
	client       http.Client
}

func init() {
	//search the home directory
	homeDir := global.GetHomeDir()
	//Create the configFilePath
	configFilePath = homeDir + "/" + configFile
}

//NewClient creates a new client
func NewClient(urlIn string, outputFormat string, timeout string, insecureSSL bool) (*ConfigManagerClient, error) {
	var c *ConfigManagerClient
	//Set the default values
	cd := &ConfigManagerClient{
		URL:          global.DefaultUrl,
		OutputFormat: global.DefaultOutputFormat,
		Timeout:      global.DefaultTimeout,
		InsecureSSL:  global.DefaultInsecureSSL,
	}
	//Search for config file
	data, errFile := ioutil.ReadFile(configFilePath)
	if errFile == nil {
		errFile = json.Unmarshal(data, &c)
	}
	//if not found use default
	if errFile != nil {
		c = cd
	}
	//overwrite with provided values
	if urlIn != "" {
		c.URL = urlIn
	}
	if outputFormat != "" {
		c.OutputFormat = outputFormat
	}
	if timeout != "" {
		timeoutI, errInt := strconv.Atoi(timeout)
		if errInt != nil {
			return nil, errInt
		}
		c.Timeout = timeoutI
	}
	//Parse the url to find the protocol
	u, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}
	c.protocol = u.Scheme
	c.InsecureSSL = insecureSSL
	//Read the certs if https and not insecure
	if c.CACertPath != "" && !c.InsecureSSL && c.protocol == "https" {
		pem, err := ioutil.ReadFile(c.CACertPath)
		if err != nil {
			log.Print(err.Error())
			return nil, err
		}
		c.rootCertPEM = pem
	}

	//Create the http client for the transport layer
	var tlsConfig *tls.Config
	if c.protocol == "https" {
		if !c.InsecureSSL {
			// create a pool of trusted certs
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(c.rootCertPEM)
			// Setup HTTPS client
			tlsConfig = &tls.Config{
				RootCAs: caCertPool,
			}
		} else {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		//Create http client with a specific timeout and transport
		c.client = http.Client{
			Timeout:   time.Second * time.Duration(c.Timeout), // Maximum of 2 secs
			Transport: transport,
		}
	} else {
		//Create http client with a specific timeout and transport
		c.client = http.Client{
			Timeout: time.Second * time.Duration(c.Timeout), // Maximum of 2 secs
		}
	}

	return c, nil
}

//Do a restcall to a given uri
func (cmc *ConfigManagerClient) RestCall(method string, baseUrl string, uri string, bodyReader io.Reader, headers map[string]string) (string, int, error) {

	//add the base url to the uri
	url := cmc.URL + baseUrl + uri

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	//Prepare the requested
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	//add Token header
	if cmc.Token != "" {
		req.Header.Set("Authorization", "Token:"+cmc.Token)
	}
	//request the request to be close after transaction...
	//if not set get socket remaning open.
	req.Close = true
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	//Execute the request
	res, getErr := cmc.client.Do(req)
	if getErr != nil {
		fmt.Printf(getErr.Error())
		return "", http.StatusInternalServerError, getErr
	}
	//Close request boddy.
	if req.Body != nil {
		req.Body.Close()
	}
	//read the response body
	body, readErr := ioutil.ReadAll(res.Body)
	//Close response boddy as already been read
	res.Body.Close()

	if readErr != nil {
		fmt.Printf(readErr.Error())
		return "", res.StatusCode, readErr
	}

	return string(body), res.StatusCode, err
}
