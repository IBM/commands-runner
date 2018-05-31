package helloWorld

import (
	"net/http"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunnerCLI/configManagerClient"
)

type MyConfigManagerClient struct {
	CMC *configManagerClient.ConfigManagerClient
}

func NewClient(urlIn string, outputFormat string, timeout string, caCertPath string, insecureSSL bool, token string) (*MyConfigManagerClient, error) {
	client, errClient := configManagerClient.NewClient(urlIn, outputFormat, timeout, caCertPath, insecureSSL, token)
	if errClient != nil {
		return nil, errClient
	}
	myClient := &MyConfigManagerClient{client}
	return myClient, nil
}

func (cmc *MyConfigManagerClient) HelloWorld() (string, error) {
	url := "myurl"
	data, _, err := cmc.CMC.RestCall(http.MethodGet, "/", url, nil, nil)
	if err != nil {
		return "", err
	}
	return data, nil
}
