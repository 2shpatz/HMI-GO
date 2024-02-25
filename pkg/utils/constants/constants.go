package constants

import (
	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// configuration file
const (
	CONFIGURATION_DIR       = "service/config"
	CONFIGURATION_FILE      = "hmi"
	CONFIGURATION_FILE_TYPE = "yaml"
	VERSION                 = "1.0.0"
	ATE_STATE_PATH          = "/shared_data/configs/ate_service_state.txt"
)

var DeviceModels = []string{RPI, COMPULAB}

const (
	// device types
	RPI      = "raspberry"
	COMPULAB = "compulab"

	CHANNEL_BUFFER_SIZE = 10

	LOWEST_PRIORITY   = math.MaxInt64
	HIGHEST_PRIORITY  = 1
	OVERRIDE_PRIORITY = 0

	// Analog GPIOs value range
	MAX_ANALOG = 1024
	MIN_ANALOG = -1024

	INFINITE_TIMEOUT  = 0
	INFINITE_DURATION = time.Duration(math.MaxInt64)
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// func PathToPackage() string {
// 	_, filename, _, _ := runtime.Caller(0)
// 	return path.Dir(filename)
// }

// func GetBroker() string {
// 	return getenv("BROKER_ADDRESS", DEFAULT_BROKER_ADDRESS)
// }

func GetServiceName() string {
	return getenv("SERVICE_NAME", "hmi")
}

// func GetBalenaApi() string {
// 	return getenv("BALENA_SUPERVISOR_ADDRESS", DEFAULT_BALENA_SUPERVISOR_API)
// }

func GetBalenaApiKey() string {
	return getenv("BALENA_SUPERVISOR_API_KEY", "")
}

// func GetHttpPort() int {
// 	port, _ := strconv.Atoi(getenv("SERVICE_HTTP_PORT", "61665"))
// 	return port
// }

func GetDeviceId() string {
	return getenv("HOSTNAME", "")
}

func GetConfigurationDir() string {
	return getenv("CONFIGURATION_DIR", CONFIGURATION_DIR)
}

func GetConfigurationFile() string {
	return getenv("CONFIGURATION_FILE", CONFIGURATION_FILE)
}

func GetDeviceModel() (string, error) {
	model, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		return "", err
	}

	// Convert bytes to string and extract the model name
	modelName := string(model)
	parts := strings.Split(modelName, " ")
	if len(parts) == 0 {
		err := fmt.Errorf("Model wasn't found")
		return "", err
	}
	modelString := strings.ToLower(strings.Trim(parts[0], " "))
	found := false
	for _, model := range DeviceModels {
		if model == modelString {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("Device Model: %s was not found in device models list", modelString)
	}
	logger.Logger.Debugf("Device Model is: %s", modelString)
	return modelString, nil

}

func IsAteExist() bool {
	exist := getenv("ATE_EXIST", "false")
	if exist == "true" {
		return true
	}
	return false

}
