package buttons

import (
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/service/events"
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/constants"
	"fmt"
	"time"

	"eos/hmi-service/pkg/utils/logger"
	"sync"

	"github.com/spf13/viper"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/service/application"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type Button struct {
	pin          gpio.PinIO
	edgeHubProxy application.EdgeHubServiceProxy
	pressTime    configs.PressTimers
	deviceModel  string
}

var buttonsMap = make(map[string]Button)
var mu sync.Mutex

func Run(edgeHubProxy application.EdgeHubServiceProxy) {
	if err := MapButtonsFromConfigs(edgeHubProxy); err != nil {
		logger.Logger.Error("Error: Failed to map Buttons")
		panic(err)
	}

	if err := startButtonsRoutine(); err != nil {
		logger.Logger.Error("Error: Failed to start buttons routine")
		panic(err)
	}
}

func (b Button) configButtonPin() error {
	logger.Logger.Debug("set button to pullup")
	if err := b.pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		return err
	}
	return nil
}

func MapButtonsFromConfigs(edgeHubProxy application.EdgeHubServiceProxy) error {
	// creates a buttons map from the loaded configurations
	ButtonsMap := GetButtonsMap()
	mu.Lock()
	defer mu.Unlock()

	logger.Logger.Debugf("Config Button map: %v", *configs.GetConfigsButtonMap())
	for _, buttonInfo := range *configs.GetConfigsButtonMap() {
		logger.Logger.Infof("Configure Button %s", buttonInfo.Alias)
		gpio := fmt.Sprintf("GPIO%d", buttonInfo.Bcm)
		model, err := constants.GetDeviceModel()
		if err != nil {
			return err
		}

		(*ButtonsMap)[string(buttonInfo.Alias)] = Button{
			pin:          gpioreg.ByName(gpio),
			edgeHubProxy: edgeHubProxy,
			pressTime:    buttonInfo.PressTimers,
			deviceModel:  model,
		}
		if (*ButtonsMap)[string(buttonInfo.Alias)].pin == nil {
			err := fmt.Errorf("Failed to find %s", gpio)
			logger.Logger.Fatal(err)
			return err
		} else {
			(*ButtonsMap)[string(buttonInfo.Alias)].configButtonPin()
		}

	}
	return nil
}
func GetButtonsMap() *map[string]Button {
	return &buttonsMap
}

func startButtonsRoutine() error {
	buttonsMap := GetButtonsMap()

	for key := range *buttonsMap {
		(*buttonsMap)[key].routineButton()
	}
	return nil
}

func (b Button) isPress() bool {
	switch {
	case b.deviceModel == constants.RPI:
		return b.pin.Read() == gpio.Low
	case b.deviceModel == constants.COMPULAB:
		return false
	default:
		// logger.Logger.Debug(b.deviceModel)
		return false
	}
}

func (b Button) routineButton() {
	var pressStartTime time.Time
	pressed := false
	shortCmdExecuted := false
	longCmdExecuted := false

	go func() {
		for {
			if b.isPress() {
				if !pressed {
					pressStartTime = time.Now()
					pressed = true
					leds.OverrideLedState(string(hmi.Cloud), string(hmi.Off), constants.INFINITE_TIMEOUT)
					logger.Logger.Debug("Button pressed")
				} else if time.Since(pressStartTime) >= time.Duration(b.pressTime.Short)*time.Second && time.Since(pressStartTime) < time.Duration(b.pressTime.Long)*time.Second {
					// Button has been pressed for a short time
					if !shortCmdExecuted {
						logger.Logger.Warnf("Button pressed for more then %d seconds", b.pressTime.Short)
						shortCmdExecuted = true
						// leds.DisableOverrideLedState(string(hmi.Cloud))
						leds.OverrideLedState(string(hmi.Cloud), string(hmi.Blink), viper.GetUint(configs.ConfigTimersOverrideOpenAp))
					}

				} else if time.Since(pressStartTime) >= time.Duration(b.pressTime.Long)*time.Second {
					// Button has been pressed for a long time
					if !longCmdExecuted {
						logger.Logger.Warnf("Button pressed for more then %d seconds", b.pressTime.Long)
						leds.OverrideAllLeds(string(hmi.Off), uint(constants.INFINITE_TIMEOUT))
						time.Sleep(200 * time.Millisecond)
						leds.OverrideAllLeds(string(hmi.Flick), uint(constants.INFINITE_TIMEOUT))
						longCmdExecuted = true
					}
				}
			} else {
				if pressed {
					// checks press duration
					releasedTime := time.Now()
					pressed = false
					elapsed := releasedTime.Sub(pressStartTime)
					elapsedInSeconds := int(elapsed / time.Second)
					shortCmdExecuted = false
					longCmdExecuted = false

					logger.Logger.Debugf("Button released after %s\n", elapsed)
					if elapsedInSeconds < int(b.pressTime.Short) {
						logger.Logger.Debug("Event: click press")
						leds.DisableOverrideLedState(string(hmi.Cloud))
						events.SendHmiButtonPressEvent(b.edgeHubProxy, string(hmi.MainButton), string(hmi.Click))
					} else if elapsedInSeconds < int(b.pressTime.Long) && elapsedInSeconds >= int(b.pressTime.Short) {
						logger.Logger.Debug("Event: short press")
						events.SendHmiButtonPressEvent(b.edgeHubProxy, string(hmi.MainButton), string(hmi.Short))
						events.SendHmiOpenWifiApEvent(b.edgeHubProxy)

					} else if elapsedInSeconds >= int(b.pressTime.Long) {
						logger.Logger.Debug("Event: long press")
						events.SendHmiButtonPressEvent(b.edgeHubProxy, string(hmi.MainButton), string(hmi.Long))
						events.SendHmiFactoryResetEvent(b.edgeHubProxy)
						leds.DisableAllOverride()
					}
				}
			}
			time.Sleep(100 * time.Millisecond) // Adjust this as needed for debounce.
		}
	}()
}
