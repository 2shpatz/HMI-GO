package gpios

import (
	"eos/hmi-service/pkg/service/events"
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/constants"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"sync"
	"time"

	"github.com/warthog618/gpiod"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/sources/sdk/edge-go-sdk.git/edge-go-sdk/service/application"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type GPIO struct {
	pin          gpio.PinIO
	direction    string
	gpioType     string
	state        int
	edgeHubProxy application.EdgeHubServiceProxy
}

var gpiosMap = make(map[string]GPIO)
var mu sync.Mutex

func setCompulabGpioSettings() error {

	type ChipSetGpios struct {
		chipSet int
		gpio    int
	}
	gpiosToSetLow := []ChipSetGpios{{5, 12}, {5, 13}, {5, 14}}

	for _, pin := range gpiosToSetLow {
		chipString := fmt.Sprintf("gpiochip%d", pin.chipSet)
		chip, err := gpiod.NewChip(chipString)
		if err != nil {
			return err
		}
		line, err := chip.RequestLine(pin.gpio, gpiod.AsOutput(0))
		if err != nil {
			return err
		}
		err = line.SetValue(0)
		if err != nil {
			return err
		}
	}

	return nil
}

func setDeviceInitialGpioSettings() error {
	modelString, err := constants.GetDeviceModel()
	if err != nil {
		logger.Logger.Errorf("Can't get the Device Model: %s", err)
	}

	switch {
	case modelString == constants.COMPULAB:
		logger.Logger.Info("Set Compulab gpio settings")
		err := setCompulabGpioSettings()
		if err != nil {
			logger.Logger.Errorf("failed to set Compulab gpio settings: %s", err)
			return err
		}
	default:
		logger.Logger.Infof("Device Model: %s has no initial GPIO settings", modelString)
	}
	return nil
}

func Run(edgeHubProxy application.EdgeHubServiceProxy) {

	err := MapGpiosFromConfigs(edgeHubProxy)
	if err != nil {
		logger.Logger.Error("Error: Failed to setup system GPIOs")
		panic(err)
	}
	setDeviceInitialGpioSettings()
	err = startGpioMonitoring()
	if err != nil {
		logger.Logger.Error("Error: Failed to start GPIO monitoring routing")
		panic(err)
	}
}
func MapGpiosFromConfigs(edgeHubProxy application.EdgeHubServiceProxy) error {
	// creates a GPIOs map from the loaded configurations
	gpiosMap := GetGpiosMap()
	mu.Lock()
	defer mu.Unlock()
	// logger.Logger.Debugf("Config GPIO map: %v", *configs.GetConfigsGPIOMap())
	for _, gpioInfo := range *configs.GetConfigsGpioMap() {
		logger.Logger.Infof("Configure GPIO %s", gpioInfo.Alias)
		gpio := fmt.Sprintf("GPIO%d", gpioInfo.Bcm)
		(*gpiosMap)[string(gpioInfo.Alias)] = GPIO{
			// pin:          gpio.NewPin(int(gpioInfo.Bcm)),
			pin:          gpioreg.ByName(gpio),
			direction:    string(gpioInfo.Direction),
			gpioType:     string(gpioInfo.Type),
			state:        gpioInfo.InitialState,
			edgeHubProxy: edgeHubProxy,
		}
		if (*gpiosMap)[string(gpioInfo.Alias)].pin == nil {
			err := fmt.Errorf("Failed to find %s", gpio)
			logger.Logger.Error(err)
			return err
		}
	}
	return nil
}

func startGpioMonitoring() error {
	inputDigitalGpios := make(map[string]GPIO)
	gpiosMap := GetGpiosMap()
	logger.Logger.Debugf("gpio map %v", gpiosMap)
	for name, gpio := range *gpiosMap {
		if gpio.direction == string(hmi.In) && gpio.gpioType == string(hmi.Digital) {
			// filter Input Digital GPIOs
			inputDigitalGpios[name] = gpio
		}
	}
	updated := false
	go func() {
		for {
			updated = false
			for alias, g := range inputDigitalGpios {
				if g.state == 1 && g.pin.Read() == gpio.Low {
					logger.Logger.Warnf("Device: %s, changed state to LOW(0)", alias)
					g.state = 0
					updated = true
				} else if g.state == 0 && g.pin.Read() == gpio.High {
					logger.Logger.Warnf("Device: %s, changed state to HIGH(1)", alias)
					g.state = 1
					updated = true
				}
				if updated {
					(*gpiosMap)[alias] = g
					inputDigitalGpios[alias] = g
					events.SendHmiGpioToggledEvent(g.edgeHubProxy, alias, g.gpioType, g.direction, g.state)
				}
			}

			time.Sleep(2 * time.Second)
		}
	}()
	return nil
}

func GetGpiosMap() *map[string]GPIO {
	return &gpiosMap
}

func GetGpio(alias string) (*GPIO, error) {
	gpiosMap := GetGpiosMap()
	value, exists := (*gpiosMap)[alias]
	if exists {
		return &value, nil
	} else {
		return &GPIO{}, fmt.Errorf("alias %s not found in gpiosMap", alias)
	}
}

func (g GPIO) GetGpioState() int {
	return g.state
}

func (g GPIO) GetGpioDirection() string {
	return g.direction
}

func (g GPIO) GetGpioType() string {
	return g.gpioType
}

func (g GPIO) highGPIO() {
	g.pin.Out(gpio.High)
}

func (g GPIO) lowGpio() {
	// g.pin.Low()
	g.pin.Out(gpio.Low)
}

func SetGpioState(alias string, state int) error {
	gpiosMap := GetGpiosMap()
	if gpio, ok := (*gpiosMap)[alias]; ok {
		if gpio.GetGpioDirection() == string(hmi.In) {
			return fmt.Errorf("Can't set any value to the input GPIO %s", alias)
		}
		if gpio.GetGpioType() == string(hmi.Digital) {
			if state != 0 && state != 1 {
				return fmt.Errorf("GPIO %s is Digital, can be set to 0/1 (not %d)", alias, state)
			}
		} else {
			if state < constants.MIN_ANALOG || state > constants.MAX_ANALOG {
				return fmt.Errorf("GPIO %s is Analog, can be set from %d to %d (not %d)", alias, constants.MIN_ANALOG, constants.MAX_ANALOG, state)
			}
		}
		gpio.state = state
		(*gpiosMap)[alias] = gpio
	} else {
		return fmt.Errorf("%s dosen't exist in GPIOs Map", alias)
	}
	return nil
}
