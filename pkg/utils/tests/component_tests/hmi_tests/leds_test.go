package commands_tests

import (
	"encoding/json"
	"eos/hmi-service/pkg/utils/constants"
	"eos/hmi-service/pkg/utils/logger"
	restlibtest "eos/hmi-service/pkg/utils/tests/restLibTest"
	"fmt"
	"testing"
	"time"

	asserts "github.com/stretchr/testify/assert"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

const (
	targetIp = "10.5.20.184"
	port     = 8080

	// priorities
	INVERTER_CONNECTED_WITH_ERRORS = 4
	INVERTER_DISCONNECTED          = 6
	INVERTER_CONNECTED             = 8
	ALL_INVERTER_CONNECTED         = 10

	PRIORITY1 = 1
	PRIORITY2 = 2
	PRIORITY3 = 3
	PRIORITY4 = 4

	// command IDs
	inverter1 = "inv1"
	inverter2 = "inv2"

	// service URIs
	uri1 = "uri1"
	uri2 = "uri2"
	uri3 = "uri3"
)

var TestLedsList = []hmi.LEDList{hmi.Power, hmi.LocalNetwork, hmi.Cloud}
var PrioritiesList = []uint{INVERTER_DISCONNECTED, ALL_INVERTER_CONNECTED}
var commandIdsList = []string{inverter1, inverter2}

func PrintMsgDebug(msg any) {
	empJSON, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		logger.Logger.Error(err)
	}
	fmt.Println(string(empJSON) + "\n")
}

// type InitLed struct{
// 	led hmi.LEDList
// 	state string
// }

func startUp() {
	restlibtest.SendReleaseAllCommands(targetIp, port, hmi.HmiServiceUri)
	restlibtest.SendReleaseAllCommands(targetIp, port, uri1)
	restlibtest.SendReleaseAllCommands(targetIp, port, uri2)
	restlibtest.SendReleaseAllCommands(targetIp, port, uri3)
}

func teardown() {

	for _, led := range TestLedsList {
		for _, priority := range PrioritiesList {
			for _, command_id := range commandIdsList {
				restlibtest.SendReleaseLed(targetIp, port, led, command_id, priority)
			}
		}
	}
}

func Test_setLedToOnAndRelease(t *testing.T) {
	t.Skip()
	assert := asserts.New(t)
	startUp()
	led_to_set := hmi.Cloud
	state_to_set := hmi.On
	command_id := inverter1
	resp := restlibtest.SendGetLed(targetIp, port, led_to_set)
	led_initial_state := resp.LEDInfo.State

	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id, state_to_set, ALL_INVERTER_CONNECTED, constants.INFINITE_TIMEOUT)
	respStruct := restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_set, fmt.Sprintf("check1: %s not equal to %s", respStruct.LEDInfo.State, state_to_set))

	state_to_check := led_initial_state
	restlibtest.SendReleaseLed(targetIp, port, led_to_set, command_id, ALL_INVERTER_CONNECTED)
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check2:%s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// teardown()
}

func Test_scenario1(t *testing.T) {
	// t.Skip()
	/*
		Shai's test request
		1. 2 inverters connect
		2. inverter1 disconnect for 30 sec set Blink
		3. inverter2 disconnect for 10 sec set Flick (Flick wins)
		4. inverter2 disconnect request expired (should Blink)
		5. inverter1 request released, back to all connect (ON)
	*/
	assert := asserts.New(t)
	startUp()

	command_id1 := inverter1
	command_id2 := inverter2
	// 1. 2 inverters connect
	led_to_set := hmi.LocalNetwork
	state_to_set := hmi.On
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id1, state_to_set, ALL_INVERTER_CONNECTED, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id2, state_to_set, ALL_INVERTER_CONNECTED, constants.INFINITE_TIMEOUT)

	//check1
	state_to_check := state_to_set
	respStruct := restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check1: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// 2. inverter1 disconnect for 30 sec set Blink
	state_to_set = hmi.Blink
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id1, state_to_set, INVERTER_DISCONNECTED, 30)

	//check2
	state_to_check = state_to_set
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check2: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// 3. inverter2 disconncet for 10 sec set Flick (Flick wins)
	state_to_set = hmi.Flick
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id2, state_to_set, INVERTER_DISCONNECTED, 10)

	//check3
	state_to_check = state_to_set
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check3: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	time.Sleep(15 * time.Second)

	// 4. inverter2 disconnect request expired (should Blink)
	// check4

	state_to_check = hmi.Blink
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check4: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// 5. inverter1 requqst released, back to all connect (ON)
	restlibtest.SendReleaseLed(targetIp, port, led_to_set, command_id1, INVERTER_DISCONNECTED)

	// check5
	state_to_check = hmi.On
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check5: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// teardown()

}

func Test_scenario2(t *testing.T) {
	// t.Skip()
	/*
		1. 2 inverters connect
		2. inverter1 disconnect
		3. inverter2 disconncet
		4. inverter1 cancel request
		5. inverter2 requqst time passes
		6. back to all connect
	*/
	assert := asserts.New(t)
	startUp()
	led_to_set := hmi.LocalNetwork
	state_to_set := hmi.On

	command_id1 := inverter1
	command_id2 := inverter2
	// inverters connected

	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id1, state_to_set, ALL_INVERTER_CONNECTED, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id2, state_to_set, ALL_INVERTER_CONNECTED, constants.INFINITE_TIMEOUT)

	//check1
	state_to_check := state_to_set
	respStruct := restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check1: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// inverter1 disconnected
	state_to_set = hmi.Blink
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id1, state_to_set, INVERTER_DISCONNECTED, 10)

	//check2
	state_to_check = state_to_set
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check2: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	time.Sleep(5 * time.Second)
	// inverter2 disconnected add more time
	state_to_set = hmi.Flick
	restlibtest.SendSetLed(targetIp, port, led_to_set, uri1, command_id2, state_to_set, INVERTER_DISCONNECTED, 10)

	//check3
	state_to_check = state_to_set
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check3: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	time.Sleep(6 * time.Second)

	//check4: inverter1 request finished, but inverter2 request still on
	state_to_check = hmi.Flick
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check4: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	//cancel inverter1 disconnected request even time has passed
	restlibtest.SendReleaseLed(targetIp, port, led_to_set, command_id1, INVERTER_DISCONNECTED)

	//check5 inverter2 disconnected request still valid
	state_to_check = hmi.Flick
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check5: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	time.Sleep(6 * time.Second)

	//check6: inverter2 request finished priority deleted
	state_to_check = hmi.On
	respStruct = restlibtest.SendGetLed(targetIp, port, led_to_set)
	assert.Equal(respStruct.LEDInfo.State, state_to_check, fmt.Sprintf("check6: %s not equal to %s", respStruct.LEDInfo.State, state_to_check))

	// teardown()

}

func Test_releaseAllCommands(t *testing.T) {
	// t.Skip()
	startUp()
	assert := asserts.New(t)
	serviceUri1 := "uri1"
	serviceUri2 := "uri2"
	serviceUri3 := "uri3"
	commandId1 := "inverter1"
	commandId2 := "ev_charger2"
	commandId3 := "heat_pump1"
	led1 := hmi.Power
	led2 := hmi.LocalNetwork
	led3 := hmi.Cloud

	resp := restlibtest.SendGetLed(targetIp, port, led1)
	led1_initial_state := resp.LEDInfo.State
	resp = restlibtest.SendGetLed(targetIp, port, led2)
	led2_initial_state := resp.LEDInfo.State
	resp = restlibtest.SendGetLed(targetIp, port, led3)
	led3_initial_state := resp.LEDInfo.State

	restlibtest.SendSetLed(targetIp, port, led1, serviceUri1, commandId1, hmi.Blink, PRIORITY1, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led1, serviceUri1, commandId1, hmi.Flick, PRIORITY2, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led2, serviceUri1, commandId1, hmi.On, PRIORITY1, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led2, serviceUri1, commandId1, hmi.Blink, PRIORITY3, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led3, serviceUri1, commandId1, hmi.On, PRIORITY1, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led3, serviceUri1, commandId1, hmi.Flick, PRIORITY2, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led1, serviceUri2, commandId2, hmi.Off, PRIORITY3, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led1, serviceUri3, commandId3, hmi.On, PRIORITY4, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led2, serviceUri3, commandId3, hmi.Blink, PRIORITY2, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led2, serviceUri2, commandId2, hmi.Off, PRIORITY4, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	restlibtest.SendSetLed(targetIp, port, led3, serviceUri2, commandId2, hmi.Blink, PRIORITY2, constants.INFINITE_TIMEOUT)
	restlibtest.SendSetLed(targetIp, port, led3, serviceUri3, commandId3, hmi.On, PRIORITY2, constants.INFINITE_TIMEOUT)
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led1)
	assert.Equal(resp.LEDInfo.State, hmi.Blink, fmt.Sprintf("check1: %s not equal to %s", resp.LEDInfo.State, hmi.Blink))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led2)
	assert.Equal(resp.LEDInfo.State, hmi.On, fmt.Sprintf("check2: %s not equal to %s", resp.LEDInfo.State, hmi.On))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led3)
	assert.Equal(resp.LEDInfo.State, hmi.On, fmt.Sprintf("check3: %s not equal to %s", resp.LEDInfo.State, hmi.On))
	time.Sleep(time.Second)
	restlibtest.SendReleaseAllCommands(targetIp, port, serviceUri1)
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led1)
	assert.Equal(resp.LEDInfo.State, hmi.Off, fmt.Sprintf("check4: %s not equal to %s", resp.LEDInfo.State, hmi.Off))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led2)
	assert.Equal(resp.LEDInfo.State, hmi.Blink, fmt.Sprintf("check5: %s not equal to %s", resp.LEDInfo.State, hmi.Blink))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led3)
	assert.Equal(resp.LEDInfo.State, hmi.Blink, fmt.Sprintf("check6: %s not equal to %s", resp.LEDInfo.State, hmi.Blink))
	time.Sleep(time.Second)
	restlibtest.SendReleaseAllCommands(targetIp, port, serviceUri2)
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led1)
	assert.Equal(resp.LEDInfo.State, hmi.On, fmt.Sprintf("check7: %s not equal to %s", resp.LEDInfo.State, hmi.On))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led2)
	assert.Equal(resp.LEDInfo.State, hmi.Blink, fmt.Sprintf("check8: %s not equal to %s", resp.LEDInfo.State, hmi.Blink))
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led3)
	assert.Equal(resp.LEDInfo.State, hmi.On, fmt.Sprintf("check9: %s not equal to %s", resp.LEDInfo.State, hmi.On))
	time.Sleep(time.Second)
	restlibtest.SendReleaseAllCommands(targetIp, port, serviceUri3)
	time.Sleep(time.Second)
	resp = restlibtest.SendGetLed(targetIp, port, led1)
	assert.Equal(resp.LEDInfo.State, led1_initial_state, fmt.Sprintf("check10: %s not equal to %s", resp.LEDInfo.State, led1_initial_state))
	resp = restlibtest.SendGetLed(targetIp, port, led2)
	assert.Equal(resp.LEDInfo.State, led2_initial_state, fmt.Sprintf("check11: %s not equal to %s", resp.LEDInfo.State, led2_initial_state))
	resp = restlibtest.SendGetLed(targetIp, port, led3)
	assert.Equal(resp.LEDInfo.State, led3_initial_state, fmt.Sprintf("check12: %s not equal to %s", resp.LEDInfo.State, led3_initial_state))
	// teardown()
}
