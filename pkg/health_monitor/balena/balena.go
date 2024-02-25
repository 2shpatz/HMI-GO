package balena

import (
	"encoding/json"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"io"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const (
	ServicesKey   = "services"
	DOWNLOAD_NONE = 0
)

type Balena struct {
	HttpClient   *http.Client
	BalenaDevice BalenaDeviceStruct
	Services     map[string]BalenaService
	BalenaApi    string
	BalenaApiKey string
}

type BalenaDeviceStruct struct {
	ApiPort           uint   `json:"api_port"`
	IpAddress         string `json:"ip_address"`
	Commit            string `json:"commit"`
	Status            string `json:"status"`
	DownloadProgress  uint   `json:"download_progress"`
	OsVersion         string `json:"os_version"`
	MacAddress        string `json:"mac_address"`
	SupervisorVersion string `json:"supervisor_version"`
	UpdatePending     bool   `json:"update_pending"`
	UpdateDownloaded  bool   `json:"update_downloaded"`
	UpdateFailed      bool   `json:"update_failed"`
}

type BalenaService struct {
	Status string `json:"status"`
	// ReleaseId        uint   `json:"releaseId"`
	DownloadProgress uint `json:"downloadProgress"`
}

type JsonDecoder struct {
	Other map[string]interface{} `mapstructure:",remain"`
}

func createGetRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Logger.Errorf("Error creating request: %s", err)
	}
	return req
}

func createPostRequest(url string, data string) {

}

func (b *Balena) CheckApiKey() error {
	if b.BalenaApiKey == "" {
		err := fmt.Errorf("ApiKey is missing, check the environment variables")
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (b *Balena) sendRequest(request *http.Request) ([]byte, error) {
	response, err := b.HttpClient.Do(request)
	if err != nil {
		logger.Logger.Errorf("Error sending request: %s", err)
		return []byte{}, err
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Logger.Errorf("Error reading response: %s", err)
		return []byte{}, err
	}
	if string(responseBody) == "Unauthorized" {
		err = fmt.Errorf("Unauthorized request")
		logger.Logger.Errorf("Error  %s", err)
		return []byte{}, err
	}
	return responseBody, nil
}

func (b *Balena) GetDevice() error {
	getDeviceUrl := fmt.Sprintf("%s/v1/device?apikey=%s", b.BalenaApi, b.BalenaApiKey)
	request := createGetRequest(getDeviceUrl)
	request.Header.Add("Content-Type", "application/json")
	responseBody, err := b.sendRequest(request)
	if err != nil {
		return err
	}

	// logger.Logger.Debugf("responseBody: %s", responseBody)
	err = json.Unmarshal(responseBody, &b.BalenaDevice)
	if err != nil {
		return err
	}

	// logger.Logger.Debugf("device status:\n%v", b.BalenaDevice)
	return nil

}

func retriveKey(key string, m map[string]interface{}) (interface{}, error) {
	if val, ok := m[key]; ok {
		return val, nil
	}
	for _, v := range m {
		if childMap, ok := v.(map[string]interface{}); ok {
			if val, _ := retriveKey(key, childMap); val != nil {
				return val, nil
			}
		}
	}
	err := fmt.Errorf("Key %s was not found", key)
	logger.Logger.Error(err)
	return nil, err
}

func (b *Balena) GetServices() error {
	var dataMap map[string]interface{}

	getServicesUrl := fmt.Sprintf("%s/v2/applications/state?apikey=%s", b.BalenaApi, b.BalenaApiKey)
	request := createGetRequest(getServicesUrl)
	responseBody, err := b.sendRequest(request)
	if err != nil {
		return err
	}
	// logger.Logger.Debugf("responseBody: %s", responseBody)
	err = json.Unmarshal(responseBody, &dataMap)
	if err != nil {
		return err
	}

	var result JsonDecoder
	targetKey := ServicesKey

	key, err := retriveKey(targetKey, dataMap)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(key, &result)
	if err != nil {
		return err
	}

	for service, info := range result.Other {
		logger.Logger.Debugf("service: %s", service)
		if innerMap, ok := info.(map[string]interface{}); ok {
			var download uint
			download = DOWNLOAD_NONE
			if innerMap["downloadProgress"] != nil {
				download = uint(innerMap["downloadProgress"].(float64))
			}

			infoStruct := BalenaService{
				DownloadProgress: download,
				Status:           innerMap["status"].(string),
			}
			b.Services[service] = infoStruct
		}

	}

	return nil
}
