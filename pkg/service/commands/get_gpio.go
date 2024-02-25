package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/gpios"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type GetGpioCommand struct {
	commandValue hmi.HmiGetGpioCommandArgs
}

func (c *GetGpioCommand) Run() (any, error) {
	var commandResponse hmi.HmiGetGpioCommandResp
	gpio, err := gpios.GetGpio(string(c.commandValue.Alias))
	if err != nil {
		return nil, err
	}

	gpioStruct := hmi.Gpio{
		Alias:     c.commandValue.Alias,
		Type:      hmi.GpioType(gpio.GetGpioType()),
		Direction: hmi.GpioDirection(gpio.GetGpioDirection()),
		State:     int64(gpio.GetGpioState()),
	}
	commandResponse.GpioInfo = gpioStruct

	return commandResponse, nil
}

func (c *GetGpioCommand) ParsedParameters(parameters map[string]interface{}) error {
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
