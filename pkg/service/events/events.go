package events

import (
	"eos/hmi-service/pkg/utils/logger"
	"time"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/service/application"
)

func SendHmiButtonPressEvent(edgeHubProxy application.EdgeHubServiceProxy, alias string, pressType string) error {
	logger.Logger.Debug("SendHmiButtonPressEvent")
	api := &hmi.HmiButtonPressedEvent{
		Timestamp: time.Now().String(),
		Alias:     hmi.ButtonList(alias),
		PressType: hmi.ButtonPressType(pressType),
	}
	// jsonData, err := json.Marshal(api)
	// if err != nil {
	// 	logger.Logger.Error("Failed to convert struct to JSON")
	// }
	// logger.Logger.Debugf("json data %s", jsonData)
	err := edgeHubProxy.PublishServiceEvent(hmi.HmiEventsButtonPressedUri, api)
	if err != nil {
		logger.Logger.Error("Error: Failed to send event: HmiButtonPressedEvent")
	}
	return nil
}

func SendHmiGpioToggledEvent(edgeHubProxy application.EdgeHubServiceProxy, alias string, gpiotype string, gpioDirection string, gpioState int) error {
	logger.Logger.Debug("SendHmiGpioToggledEvent")
	api := hmi.HmiGpioToggledEvent{
		Timestamp: time.Now().String(),
		GpioInfo: hmi.Gpio{
			Alias:     hmi.GpioList(alias),
			Type:      hmi.GpioType(gpiotype),
			Direction: hmi.GpioDirection(gpioDirection),
			State:     int64(gpioState),
		},
	}
	err := edgeHubProxy.PublishServiceEvent(hmi.HmiEventsGpioToggledUri, api)
	if err != nil {
		logger.Logger.Error("Failed to send event: HmiGpioToggledEvent")
	}
	return nil

}

func SendHmiOpenWifiApEvent(edgeHubProxy application.EdgeHubServiceProxy) error {
	logger.Logger.Debug("SendHmiOpenWifiApEvent")
	api := hmi.HmiOpenWifiApEvent{
		Timestamp: time.Now().String(),
	}
	err := edgeHubProxy.PublishServiceEvent(hmi.HmiEventsOpenWifiApUri, api)
	if err != nil {
		logger.Logger.Error("Failed to send event: HmiOpenWifiApEvent")
	}
	return nil
}

func SendHmiFactoryResetEvent(edgeHubProxy application.EdgeHubServiceProxy) error {
	logger.Logger.Debug("SendHmiOpenWifiApEvent")
	api := hmi.HmiFactoryResetEvent{
		Timestamp: time.Now().String(),
	}
	err := edgeHubProxy.PublishServiceEvent(hmi.HmiEventsFactoryResetUri, api)
	if err != nil {
		logger.Logger.Error("Failed to send event: HmiFactoryResetEvent")
	}
	return nil
}
