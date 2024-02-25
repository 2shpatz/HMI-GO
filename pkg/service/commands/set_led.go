package commands

import (
	"encoding/json"
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/utils/constants"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type SetLedCommand struct {
	commandValue hmi.HmiSetLEDCommandArgs
}

func (c *SetLedCommand) Run() (any, error) {

	alias := string(c.commandValue.Alias)
	commandId := string(c.commandValue.CommandID)
	var serviceUri string
	state := string(c.commandValue.State)
	priority := uint(c.commandValue.Priority)
	timeout := uint(c.commandValue.Timeout)

	var commandResponse hmi.HmiSetLEDCommandResp
	if alias == "" {
		err := fmt.Errorf("Alias field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}
	logger.Logger.Debugf("c.commandValue.ServiceURI %s", *c.commandValue.ServiceURI)
	if c.commandValue.ServiceURI == nil {
		serviceUri = ""
	} else {
		serviceUri = *c.commandValue.ServiceURI
	}
	if commandId == "" {
		err := fmt.Errorf("CommandId field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}
	if state == "" {
		err := fmt.Errorf("State field is empty discarding request")
		logger.Logger.Error(err)
		return commandResponse, err
	}
	if priority < constants.HIGHEST_PRIORITY {
		err := fmt.Errorf("Priority was not set or lower then %d, select priority %d or greater", constants.HIGHEST_PRIORITY, constants.HIGHEST_PRIORITY)
		logger.Logger.Error(err)
		return commandResponse, err
	}
	if timeout < 0 {
		err := fmt.Errorf("Request timeout is: %d, please set to [timeout >= 0]", c.commandValue.Timeout)
		logger.Logger.Error(err)
		return commandResponse, err
	}
	priorityKey := leds.PriorityMapKey{
		Level:     priority,
		CommandId: commandId,
	}
	logger.Logger.Debugf("Run: SetLedCommand, with parameters: alias: %s, command_id: %s, state: %s, priority: %d, timeout: %d", alias, commandId, state, priority, timeout)
	err := leds.AddPriorityListing(alias, priorityKey, state, timeout, serviceUri)

	return commandResponse, err
}

func (c *SetLedCommand) ParsedParameters(parameters map[string]interface{}) error {
	c.commandValue = hmi.HmiSetLEDCommandArgs{}
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

func (c *SetLedCommand) SetCommandValue(parameters hmi.HmiSetLEDCommandArgs) {
	c.commandValue = parameters
}
