package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/gpios"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type SetGpioCommand struct {
	commandValue hmi.HmiSetGpioCommandArgs
}

func (c *SetGpioCommand) Run() (any, error) {
	err := gpios.SetGpioState(string(c.commandValue.Alias), int(c.commandValue.State))

	commandResponse := hmi.HmiSetGpioCommandResp{}
	return commandResponse, err
}

func (c *SetGpioCommand) ParsedParameters(parameters map[string]interface{}) error {
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
