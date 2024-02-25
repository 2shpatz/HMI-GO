package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/buttons"
	"eos/hmi-service/pkg/hmi/gpios"
	"eos/hmi-service/pkg/hmi/leds"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type GetAllGpiosCommand struct {
	commandValue hmi.HmiGetLEDCommandArgs
}

func convertLedsMapToList(mapToConvert map[string]leds.LED) []hmi.LED {
	var list []hmi.LED
	for key, value := range mapToConvert {
		ledStruct := hmi.LED{
			Alias: hmi.LEDList(key),
			State: hmi.LEDStates(value.GetLedState()),
		}
		list = append(list, ledStruct)
	}
	return list
}

func convertGpiosMapToList(mapToConvert map[string]gpios.GPIO) []hmi.Gpio {
	var list []hmi.Gpio
	for key, value := range mapToConvert {
		gpioStruct := hmi.Gpio{
			Alias:     hmi.GpioList(key),
			Type:      hmi.GpioType(value.GetGpioType()),
			Direction: hmi.GpioDirection(value.GetGpioDirection()),
			State:     int64(value.GetGpioState()),
		}
		list = append(list, gpioStruct)
	}
	return list
}

func convertButtonsMapToList(mapToConvert map[string]buttons.Button) []hmi.ButtonList {
	var list []hmi.ButtonList
	for key := range mapToConvert {
		list = append(list, hmi.ButtonList(key))
	}
	return list
}

func (c *GetAllGpiosCommand) Run() (any, error) {
	ledsMap := leds.GetLedsMap()
	gpiosMap := gpios.GetGpiosMap()
	battonsMap := buttons.GetButtonsMap()

	ledsList := convertLedsMapToList(*ledsMap)
	gpiosList := convertGpiosMapToList(*gpiosMap)
	buttonsList := convertButtonsMapToList(*battonsMap)

	commandResponse := hmi.HmiGetAllGpiosCommandResp{
		ButtonAlias: buttonsList,
		GpioInfo:    gpiosList,
		LEDInfo:     ledsList,
	}

	return commandResponse, nil
}

func (c *GetAllGpiosCommand) ParsedParameters(parameters map[string]interface{}) error {
	data, err := json.Marshal(parameters)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &c.commandValue)
	if err != nil {
		return err
	}
	return nil

}
