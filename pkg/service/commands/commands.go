package commands

import (
	"fmt"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type ServiceCommand interface {
	ParsedParameters(parameters map[string]interface{}) error
	Run() (any, error)
}

func GetCommand(uri string) (ServiceCommand, error) {
	if category, ok := serviceCommandsMap[uri]; ok {
		return category, nil
	} else {
		return nil, fmt.Errorf("Command %s, not found in serviceCommandsMap", uri)
	}
}

// Used to map a The URI to the update function
var serviceCommandsMap = map[string]ServiceCommand{
	hmi.HmiCommandsSetLedUri:             &SetLedCommand{},
	hmi.HmiCommandsReleaseLedUri:         &ReleaseLedCommand{},
	hmi.HmiCommandsReleaseAllCommandsUri: &ReleaseAllCommandsCommand{},
	hmi.HmiCommandsSetGpioUri:            &SetGpioCommand{},
	hmi.HmiCommandsGetLedUri:             &GetLedCommand{},
	hmi.HmiCommandsGetGpioUri:            &GetGpioCommand{},
	hmi.HmiCommandsGetAllGpiosUri:        &GetAllGpiosCommand{},
}
