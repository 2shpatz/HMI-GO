package main

import (
	"eos/hmi-service/pkg/service"
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/logger"
	"os/exec"
	"time"

	"github.com/spf13/viper"
)

var debug = true

func runRpc() {
	go func() {
		time.Sleep(2 * time.Second)
		cmd := exec.Command("./rpc/edge-proxy-sdk", service.GetServiceName(), viper.GetString(configs.ConfigServiceRpcHttpPort), service.GetServiceVersion())
		output, err := cmd.Output()
		if err != nil {
			logger.Logger.Error("Error: " + err.Error())
		}
		logger.Logger.Debugf("rpc: %s", output)
	}()
}

func main() {
	// configure Constants
	err := configs.InitConfigs()

	logger.InitLogger(viper.GetString(configs.ConfigServiceLogLevel))
	if err != nil {
		logger.Logger.Fatalf("cannot load config: %s", err)
	}
	// Configure Service
	server, mqttClient, err := service.RunHmiService()
	defer mqttClient.Stop()
	if err != nil {
		logger.Logger.Fatalf("Can't run HMI service, error %s", err)
	}
	defer server.Stop()
	runRpc()

	server.Await()

}
