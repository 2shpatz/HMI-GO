package configs

import (
	"eos/hmi-service/pkg/utils/constants"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

const (
	// SERVICE
	ConfigServiceSupervisorAddress = "service.apis.supervisor.address"
	ConfigServiceSupervisorApiKey  = "service.apis.supervisor.api_key"
	ConfigServiceRpcHttpPort       = "service.rpc.http_port"
	ConfigServiceBrokerAddress     = "service.broker.address"
	ConfigServiceLogLevel          = "service.logs.level"

	// HARDWARE

	// LEDs
	ConfigHardwareInterfaceLedsAlias        = "hardware.interface.leds.%s.alias"
	ConfigHardwareInterfaceLedsBcm          = "hardware.interface.leds.%s.gpio_bcm"
	ConfigHardwareInterfaceLedsInitialState = "hardware.interface.leds.%s.initial_state"
	ConfigHardwareInterfaceLedsOsPath       = "hardware.interface.leds.%s.os_path"

	// Buttons
	ConfigHardwareInterfaceButtonsAlias      = "hardware.interface.buttons.%s.alias"
	ConfigHardwareInterfaceButtonsBcm        = "hardware.interface.buttons.%s.gpio_bcm"
	ConfigHardwareInterfaceButtonsPressShort = "hardware.interface.buttons.%s.press_timers.short"
	ConfigHardwareInterfaceButtonsPressLong  = "hardware.interface.buttons.%s.press_timers.long"

	// GPIOs
	ConfigHardwareInterfaceGpiosAlias        = "hardware.interface.gpios.%s.alias"
	ConfigHardwareInterfaceGpiosBcm          = "hardware.interface.gpios.%s.gpio_bcm"
	ConfigHardwareInterfaceGpiosDirection    = "hardware.interface.gpios.%s.direction"
	ConfigHardwareInterfaceGpiosType         = "hardware.interface.gpios.%s.type"
	ConfigHardwareInterfaceGpiosInitialState = "hardware.interface.gpios.%s.initial_state"
	ConfigHardwareInterfaceGpiosOsPath       = "hardware.interface.gpios.%s.os_path"

	// TIMERS
	ConfigTimersOverrideOpenAp = "timers.led_overrides.open_ap"
	ConfigTimersFlickUp        = "timers.led_states.flick_up"
	ConfigTimersFlickDown      = "timers.led_states.flick_down"
	ConfigTimersBlinkUp        = "timers.led_states.Blink_up"
	ConfigTimersBlinkDown      = "timers.led_states.Blink_down"

	// device_uuid_var_name = "BALENA_DEVICE_UUID"
)

var HardwareConfig HardwareKey
var ServiceConfig ServiceKey
var TimersConfig TimerKey

var ConfigsLedMap map[string]LedsInfo
var ConfigsButtonMap map[string]ButtonsInfo
var ConfigsGpioMap map[string]GpiosInfo
var ConfigsApiMap map[string]ApisConfig

func GetConfigsLedMap() *map[string]LedsInfo {
	return &ConfigsLedMap
}
func GetConfigsButtonMap() *map[string]ButtonsInfo {
	return &ConfigsButtonMap
}
func GetConfigsGpioMap() *map[string]GpiosInfo {
	return &ConfigsGpioMap
}

func GetConfigsApiMap() *map[string]ApisConfig {
	return &ConfigsApiMap
}

type HardwareKey struct {
	Interfaces InterfaceConfig `mapstructure:"interface"`
}

type InterfaceConfig struct {
	Leds    map[string]LedsInfo    `mapstructure:"leds"`
	Buttons map[string]ButtonsInfo `mapstructure:"buttons"`
	Gpios   map[string]GpiosInfo   `mapstructure:"gpios"`
}

type LedsInfo struct {
	Alias        hmi.LEDList `mapstructure:"alias"`
	Bcm          uint        `mapstructure:"gpio_bcm"`
	InitialState string      `mapstructure:"initial_state"`
	OsPath       string      `mapstructure:"os_path"`
}

type PressTimers struct {
	Short uint `mapstructure:"short"`
	Long  uint `mapstructure:"long"`
}

type ButtonsInfo struct {
	Alias       hmi.ButtonList `mapstructure:"alias"`
	Bcm         uint           `mapstructure:"gpio_bcm"`
	PressTimers PressTimers    `mapstructure:"press_timers"`
	OsPath      string         `mapstructure:"os_path"`
}

type GpiosInfo struct {
	Alias        hmi.GpioList      `mapstructure:"alias"`
	Bcm          uint              `mapstructure:"gpio_bcm"`
	Direction    hmi.GpioDirection `mapstructure:"direction"`
	Type         hmi.GpioType      `mapstructure:"type"`
	InitialState int               `mapstructure:"initial_state"`
	OsPath       string            `mapstructure:"os_path"`
}

type ServiceKey struct {
	Apis   map[string]ApisConfig `mapstructure:"apis"`
	Broker BrokerConfig          `mapstructure:"broker"`
	Logs   LogsConfig            `mapstructure:"logs"`
	Rpc    RpcConfig             `mapstructure:"rpc"`
}

type ApisConfig struct {
	Address string `mapstructure:"address"`
	ApiKey  string `mapstructure:"api_key"`
}

type BrokerConfig struct {
	Address string `mapstructure:"address"`
}

type LogsConfig struct {
	Level string `mapstructure:"level"`
}

type RpcConfig struct {
	HttpPort uint `mapstructure:"http_port"`
}

type TimerKey struct {
	LedOverrideDuration LedOverrideTimers `mapstructure:"led_override_duration"`
	LedStates           LedStatesTimers   `mapstructure:"led_states_timers"`
}

type LedOverrideTimers struct {
	OpenAp uint `mapstructure:"open_ap"`
}

type LedStatesTimers struct {
	FlickUp   uint `mapstructure:"flick_up"`
	FlickDown uint `mapstructure:"flick_down"`
	BlinkUp   uint `mapstructure:"blink_up"`
	BlinkDown uint `mapstructure:"blink_down"`
}

func setDefault() {
	viper.BindEnv(ConfigServiceRpcHttpPort, "SERVICE_HTTP_PORT")
	viper.SetDefault(ConfigServiceRpcHttpPort, 61665)
	viper.BindEnv(ConfigServiceLogLevel, "SERVICE_LOGS_LEVEL")
	viper.SetDefault(ConfigServiceLogLevel, "DEBUG")
	viper.BindEnv(ConfigServiceBrokerAddress, "BROKER_ADDRESS")
	viper.SetDefault(ConfigServiceBrokerAddress, "mqtt:1883")
	viper.BindEnv(ConfigServiceSupervisorAddress, "BALENA_SUPERVISOR_ADDRESS")
	viper.SetDefault(ConfigServiceSupervisorAddress, "http://127.0.0.1:48484/v2/local/device")
	viper.BindEnv(ConfigServiceSupervisorApiKey, "BALENA_SUPERVISOR_API_KEY")

	viper.SetDefault(ConfigTimersOverrideOpenAp, 15)
	viper.SetDefault(ConfigTimersFlickUp, 125)
	viper.SetDefault(ConfigTimersFlickDown, 125)
	viper.SetDefault(ConfigTimersBlinkUp, 1000)
	viper.SetDefault(ConfigTimersBlinkDown, 1000)
}

func InitConfigs() error {
	err := loadConfigs()
	if err != nil {
		return err
	}
	return nil
}

func loadConfigs() error {
	// Loads a config file that specified in constants.CONFIGURATION_DIR + SERVICE_NAME (from env variable) and creates the variables
	viper.SetConfigName(constants.GetConfigurationFile())
	viper.SetConfigType(constants.CONFIGURATION_FILE_TYPE)
	viper.AddConfigPath(constants.GetConfigurationDir())
	setDefault()
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	if err := viper.UnmarshalKey("hardware", &HardwareConfig); err != nil {
		panic(fmt.Errorf("unable to decode Hardware config: %s", err))
	}
	configLedMap := GetConfigsLedMap()
	*configLedMap = HardwareConfig.Interfaces.Leds

	configButtonMap := GetConfigsButtonMap()
	*configButtonMap = HardwareConfig.Interfaces.Buttons

	configGpioMap := GetConfigsGpioMap()
	*configGpioMap = HardwareConfig.Interfaces.Gpios

	if err := viper.UnmarshalKey("service", &ServiceConfig); err != nil {
		panic(fmt.Errorf("unable to decode Service config: %s", err))
	}

	configApisMap := GetConfigsApiMap()
	*configApisMap = ServiceConfig.Apis

	if err := viper.UnmarshalKey("timers", &TimersConfig); err != nil {
		panic(fmt.Errorf("unable to decode Timers config: %s", err))
	}

	return nil
}

func WriteTemplateConfigs() {
	viper.SetConfigName(constants.CONFIGURATION_FILE)
	viper.SetConfigType(constants.CONFIGURATION_FILE_TYPE)
	viper.AddConfigPath(constants.CONFIGURATION_DIR)

	ledBcmMap := make(map[hmi.LEDList]LedsInfo)
	ledBcmMap[hmi.Power] = LedsInfo{
		Bcm:          22,
		InitialState: string(hmi.Blink),
	}
	ledBcmMap[hmi.LocalNetwork] = LedsInfo{
		Bcm:          27,
		InitialState: string(hmi.Off),
	}
	ledBcmMap[hmi.Cloud] = LedsInfo{
		Bcm:          10,
		InitialState: string(hmi.Off),
	}

	ButtonBcmMap := make(map[hmi.ButtonList]ButtonsInfo)
	ButtonBcmMap[hmi.MainButton] = ButtonsInfo{
		Bcm: 17,
		PressTimers: PressTimers{
			Short: 4,
			Long:  10,
		},
	}

	GpioBcmMap := make(map[hmi.GpioList]GpiosInfo)
	GpioBcmMap[hmi.Generator] = GpiosInfo{
		Bcm:          4,
		Direction:    hmi.In,
		Type:         hmi.Digital,
		InitialState: 0,
	}

	for ledName, ledInfo := range ledBcmMap {
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceLedsAlias, ledName), ledName)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceLedsBcm, ledName), ledInfo.Bcm)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceLedsInitialState, ledName), ledInfo.InitialState)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceLedsOsPath, ledName), fmt.Sprintf("/sys/class/leds/gpio%d/", ledInfo.Bcm))
	}

	for buttonName, buttonInfo := range ButtonBcmMap {
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceButtonsAlias, buttonName), buttonName)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceButtonsBcm, buttonName), buttonInfo.Bcm)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceButtonsPressShort, buttonName), buttonInfo.PressTimers.Short)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceButtonsPressLong, buttonName), buttonInfo.PressTimers.Long)
	}

	for gpioName, gpioInfo := range GpioBcmMap {
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosAlias, gpioName), gpioName)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosBcm, gpioName), gpioInfo.Bcm)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosDirection, gpioName), gpioInfo.Direction)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosType, gpioName), gpioInfo.Type)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosInitialState, gpioName), gpioInfo.InitialState)
		viper.SetDefault(fmt.Sprintf(ConfigHardwareInterfaceGpiosOsPath, gpioName), fmt.Sprintf("/sys/class/gpio/gpio%d/", gpioInfo.Bcm))
	}

	setDefault()
	err := viper.WriteConfigAs(
		filepath.Join(
			constants.CONFIGURATION_DIR,
			constants.GetServiceName()+"."+constants.CONFIGURATION_FILE_TYPE,
		),
	)
	if err != nil {
		fmt.Printf("Couldn't write config file: %s", err)
	}

	fmt.Println("Writing new configs file")
	err = viper.SafeWriteConfig() // WriteConfigAs(constants.CONFIGURATION_FILE_PATH)

	viper.WriteConfig()
}
