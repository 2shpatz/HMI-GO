package restlibtest

import (
	"encoding/json"
	"fmt"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type httpMessage struct {
	ServiceUri string                 `json:"serviceUri"`
	CommandUri string                 `json:"commandUri"`
	Parameters map[string]interface{} `json:"parameters"`
}

func convertMapInterfaceToGetLedStruct(data map[string]interface{}) hmi.HmiGetLEDCommandResp {
	var getLedStruct hmi.HmiGetLEDCommandResp
	jsonValue, _ := json.Marshal(data)
	err := json.Unmarshal(jsonValue, &getLedStruct)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return getLedStruct

}

func SendSetLed(address string, port uint, alias hmi.LEDList, serviceUri string, commandId string, state hmi.LEDStates, priority uint, timeout uint) (map[string]interface{}, error) {
	var serviceUriPnt *string
	serviceUriPnt = &serviceUri
	println(serviceUriPnt)
	int64Value := int64(priority)
	if serviceUri == "" {
		serviceUri = hmi.HmiServiceUri
	}
	payload := hmi.HmiSetLEDCommandArgs{
		Alias:      alias,
		CommandID:  commandId,
		ServiceURI: serviceUriPnt,
		State:      state,
		Priority:   int64Value,
		Timeout:    int64(timeout),
	}

	req := httpMessage{
		ServiceUri: hmi.HmiServiceUri,
		CommandUri: hmi.HmiCommandsSetLedUri,
		Parameters: convertStructIntoMapInterface(payload),
	}
	fmt.Printf("sending request: %v", req)
	return sendServiceCmd(address, port, req)
}

func SendReleaseLed(address string, port uint, alias hmi.LEDList, commandId string, priority uint) (map[string]interface{}, error) {
	int64Value := int64(priority)
	payload := hmi.HmiReleaseLEDCommandArgs{
		Alias:     alias,
		CommandID: commandId,
		Priority:  int64Value,
	}

	req := httpMessage{
		ServiceUri: hmi.HmiServiceUri,
		CommandUri: hmi.HmiCommandsReleaseLedUri,
		Parameters: convertStructIntoMapInterface(payload),
	}
	fmt.Printf("sending request: %v", req)
	return sendServiceCmd(address, port, req)
}

func SendGetLed(address string, port uint, alias hmi.LEDList) hmi.HmiGetLEDCommandResp {
	payload := hmi.HmiGetLEDCommandArgs{
		Alias: alias,
	}
	req := httpMessage{
		ServiceUri: hmi.HmiServiceUri,
		CommandUri: hmi.HmiCommandsGetLedUri,
		Parameters: convertStructIntoMapInterface(payload),
	}
	fmt.Printf("sending request: %v", req)
	resp, _ := sendServiceCmd(address, port, req)
	return convertMapInterfaceToGetLedStruct(resp)
}

func SendReleaseAllCommands(address string, port uint, serviceUri string) hmi.HmiGetLEDCommandResp {
	payload := hmi.HmiReleaseAllCommandsCommandArgs{
		ServiceURI: serviceUri,
	}
	req := httpMessage{
		ServiceUri: hmi.HmiServiceUri,
		CommandUri: hmi.HmiCommandsReleaseAllCommandsUri,
		Parameters: convertStructIntoMapInterface(payload),
	}
	fmt.Printf("sending request: %v", req)
	resp, _ := sendServiceCmd(address, port, req)
	return convertMapInterfaceToGetLedStruct(resp)
}
