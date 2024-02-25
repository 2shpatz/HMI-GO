package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type ReleaseAllCommandsCommand struct {
	commandValue hmi.HmiReleaseAllCommandsCommandArgs
}

func (c *ReleaseAllCommandsCommand) Run() (any, error) {
	var commandResponse hmi.HmiReleaseAllCommandsCommandResp
	serviceUri := c.commandValue.ServiceURI

	if serviceUri == "" {
		err := fmt.Errorf("ServiceURI field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}

	logger.Logger.Debugf("Run: ReleaseAllCommandsCommand, with parameters: serviceUri: %s", serviceUri)

	err := leds.ClearServiceUriCommands(serviceUri)
	if err != nil {
		logger.Logger.Error(err)
	}
	return commandResponse, err
}

func (c *ReleaseAllCommandsCommand) ParsedParameters(parameters map[string]interface{}) error {
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

func (c *ReleaseAllCommandsCommand) SetCommandValue(parameters hmi.HmiReleaseAllCommandsCommandArgs) {
	c.commandValue = parameters
}
