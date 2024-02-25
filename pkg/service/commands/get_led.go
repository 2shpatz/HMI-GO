package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/leds"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type GetLedCommand struct {
	commandValue hmi.HmiGetLEDCommandArgs
}

func (c *GetLedCommand) Run() (any, error) {
	led, err := leds.GetLed(string(c.commandValue.Alias))
	if err != nil {
		return nil, err
	}
	ledState := led.GetLedState()

	ledStruct := hmi.LED{
		Alias: c.commandValue.Alias,
		State: hmi.LEDStates(ledState),
	}

	commandResponse := hmi.HmiGetLEDCommandResp{
		LEDInfo: ledStruct,
	}
	return commandResponse, nil
}

func (c *GetLedCommand) ParsedParameters(parameters map[string]interface{}) error {
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

func (c *GetLedCommand) SetCommandValue(parameters hmi.HmiGetLEDCommandArgs) {
	c.commandValue = parameters
}
