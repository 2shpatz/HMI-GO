package service

import (
	"encoding/json"
	"eos/hmi-service/pkg/health_monitor/apps_monitor"
	"eos/hmi-service/pkg/hmi/buttons"
	"eos/hmi-service/pkg/hmi/gpios"
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/service/commands"
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/constants"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/commands/edge_commands"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/app"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/mqtt"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/service/application"
	"periph.io/x/host/v3"
)

type hmiApplicationService struct {
	edgeHubProxy application.EdgeHubServiceProxy
}

func (m *hmiApplicationService) ServiceUri() string {
	return hmi.HmiServiceUri
}

func (m *hmiApplicationService) ServiceVersion() string {
	return GetServiceVersion() // Provided by your code
}

func (m *hmiApplicationService) RpcKeepalive(
	args edge_commands.RPCServiceKeepaliveCommandArgs) (edge_commands.RPCServiceKeepaliveCommandResp, error) {
	// log.Printf("RpcKeepalive - args %v", args)

	return edge_commands.RPCServiceKeepaliveCommandResp{BootTime: "1.31", ProgrammingLanguage: edge_commands.Go,
		SwVersion: "1.2.3"}, nil
}

func (m *hmiApplicationService) RegisterService(
	args edge_commands.RegisterServiceCommandArgs) (edge_commands.RegisterServiceCommandResp, error) {
	log.Printf("RegisterService - args %v", args)

	return edge_commands.RegisterServiceCommandResp{Version: "1.31"}, nil
}

func (m *hmiApplicationService) GetServiceProperties(
	args edge_commands.GetServicePropertiesCommandArgs) (edge_commands.GetServicePropertiesCommandResp, error) {

	return edge_commands.GetServicePropertiesCommandResp{}, nil
}

func (m *hmiApplicationService) GetServiceSettings(args edge_commands.GetServiceSettingsCommandArgs) (edge_commands.GetServiceSettingsCommandResp, error) {
	logger.Logger.Printf("GetServiceSettings - args %v", args)

	return edge_commands.GetServiceSettingsCommandResp{}, nil
}

func (m *hmiApplicationService) TriggerPublishServiceState(
	args edge_commands.TriggerPublishServiceStateCommandArgs) (edge_commands.TriggerPublishServiceStateCommandResp, error) {
	log.Printf("TriggerPublishServiceState - args %v", args)

	return edge_commands.TriggerPublishServiceStateCommandResp{}, nil
}

func (m *hmiApplicationService) SampleServiceTelemetry(
	args edge_commands.SampleServiceTelemetryCommandArgs) (edge_commands.SampleServiceTelemetryCommandResp, error) {
	log.Printf("SampleServiceTelemetry - args %v", args)

	return edge_commands.SampleServiceTelemetryCommandResp{}, nil
}

func (m *hmiApplicationService) UpdateServiceSettings(args edge_commands.UpdateServiceSettingsCommandArgs) (edge_commands.UpdateServiceSettingsCommandResp, error) {
	log.Printf("UpdateServiceSettings - args %v", args)

	resp := edge_commands.UpdateServiceSettingsCommandResp{}
	msg := "It's alright"
	for settingName := range args.SettingCategories {
		resp.SettingCategories[settingName] =
			edge_commands.SettingCategoryUpdateReply{Status: "OK", Message: &msg}
	}

	return resp, nil
}

func (m *hmiApplicationService) ExecuteServiceCommand(
	args edge_commands.ExecuteServiceCommandArgs) (edge_commands.ExecuteServiceCommandResp, error) {
	log.Printf("ExecuteServiceCommand - args %v", args)

	logger.Logger.Infof("Received New command %v", args)
	resp := edge_commands.ExecuteServiceCommandResp{}
	cmd, err := commands.GetCommand(args.CommandURI)
	if err != nil {
		logger.Logger.Errorf("Error in execute command: %s. Error: %s", args.CommandURI, err)
		return resp, err
	}
	logger.Logger.Debugf("execute command %v", args.Parameters)
	cmd.ParsedParameters(args.Parameters)

	runResp, err := cmd.Run()
	if err != nil {
		logger.Logger.Errorf("Error in execute command: %s. Error: %s", args.CommandURI, err)
		return resp, err
	}
	resp = edge_commands.ExecuteServiceCommandResp{
		Result: convertStructIntoMapInterface(runResp),
	}
	logger.Logger.Debugf("Command %s Response %v", args.CommandURI, resp)
	return resp, nil
}

func convertStructIntoMapInterface(runResp any) map[string]interface{} {
	var respFields map[string]interface{}
	inrec, _ := json.Marshal(runResp)
	json.Unmarshal(inrec, &respFields)
	return respFields
}

func GetServiceName() string {
	return hmi.HmiServiceUri // Provided by the Metadata
}

func GetServiceVersion() string {
	return constants.VERSION
}

func NewHmiApplicationService(edgeHubProxy application.EdgeHubServiceProxy) application.AppServiceInterface {
	return &hmiApplicationService{edgeHubProxy: edgeHubProxy}
}

func haltIfAteExist() {
	if constants.IsAteExist() {
		content, err := os.ReadFile(constants.ATE_STATE_PATH)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Logger.Warn("ATE state file isn't exist, halting the service until the next reboot")
				select {}
			} else {
				logger.Logger.Error(err)
				logger.Logger.Warn("Can't read ATE state file, halting the service until the next reboot")
				select {}
			}
		}
		processedContent := strings.ReplaceAll(string(content), "\n", "")
		processedContent = strings.TrimSpace(processedContent)
		if processedContent != "0" {
			logger.Logger.Warnf("ATE state file content is %s ATE is active, halting the service until the next reboot", string(content))
			select {}
		}
		logger.Logger.Infof("ATE state file content is %s ATE is down", processedContent)

	}

}

func RunHmiService() (app.AppServiceServer, mqtt.Client, error) {
	logger.Logger.Infof("[main] Starting %s, version %s\n", GetServiceName(), GetServiceVersion())

	mqttClient := mqtt.NewPahoMqttClient(viper.GetString(configs.ConfigServiceBrokerAddress), fmt.Sprint(GetServiceName()+"service"))
	err := mqttClient.Start()

	edgeHubProxy := app.NewServiceProxy(GetServiceName(), mqttClient)
	hmiService := NewHmiApplicationService(edgeHubProxy)
	server := app.NewAppServiceServer(hmiService, viper.GetInt(configs.ConfigServiceRpcHttpPort), true)
	err = server.Start()
	if err != nil {
		return server, mqttClient, err
	}
	if _, err := host.Init(); err != nil {
		logger.Logger.Fatal(err)
		return nil, nil, err
	}
	gpios.Run(edgeHubProxy)
	leds.Run()
	buttons.Run(edgeHubProxy)

	apps_monitor.Run() //to be moved to other service

	return server, mqttClient, nil
}
