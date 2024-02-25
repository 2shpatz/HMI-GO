package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/utils/constants"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type ReleaseLedCommand struct {
	commandValue hmi.HmiReleaseLEDCommandArgs
}

func (c *ReleaseLedCommand) Run() (any, error) {
	var commandResponse hmi.HmiReleaseLEDCommandResp

	alias := string(c.commandValue.Alias)
	commandId := string(c.commandValue.CommandID)
	priority := uint(c.commandValue.Priority)

	if alias == "" {
		err := fmt.Errorf("Alias field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}
	if commandId == "" {
		err := fmt.Errorf("CommandId field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}
	if priority < constants.HIGHEST_PRIORITY {
		err := fmt.Errorf("Priority was not set or lower then %d, select priority %d or greater", constants.HIGHEST_PRIORITY, constants.HIGHEST_PRIORITY)
		logger.Logger.Error(err)
		return commandResponse, nil
	}

	logger.Logger.Debugf("Run: ReleaseLedCommand, with parameters: alias: %s, command_id: %s, priority: %d", alias, commandId, priority)
	priorityKey := leds.PriorityMapKey{
		Level:     priority,
		CommandId: commandId,
	}
	err := leds.RemovePriorityListing(string(c.commandValue.Alias), priorityKey)
	if err != nil {
		logger.Logger.Error(err)
	}
	return commandResponse, err
}

func (c *ReleaseLedCommand) ParsedParameters(parameters map[string]interface{}) error {
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

func (c *ReleaseLedCommand) SetCommandValue(parameters hmi.HmiReleaseLEDCommandArgs) {
	c.commandValue = parameters
}
