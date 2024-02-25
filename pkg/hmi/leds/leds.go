package leds

import (
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/constants"
	"os"
	"sync"

	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

type PriorityMapKey struct {
	Level     uint
	CommandId string
}

type Priority struct {
	state           string
	priorityEndTime time.Time // The time to release/delete the priority
	serviceUri      string
}
type LED struct {
	path string
	// pin           *gpio.Pin
	initialState  string
	priorityMap   map[PriorityMapKey]Priority
	newPriorityCh chan Priority
}

var ledStateHierarchy = []hmi.LEDStates{hmi.Off, hmi.On, hmi.Blink, hmi.Flick} // higher index -> higher hierarchy

var ledsMap = make(map[string]LED)

var lowestPriorityEntry = Priority{
	state:           string(hmi.Flick),
	priorityEndTime: time.Now().Add(constants.INFINITE_DURATION),
}

var mutex = sync.Mutex{}

func Run() {
	err := MapLedsFromConfigs()
	if err != nil {
		logger.Logger.Error("Error: Failed to setup system LEDs")
		panic(err)
	}
	err = initLeds()
	if err != nil {
		logger.Logger.Error("Error: Failed to setup system LEDs")
		panic(err)
	}
}

func initLeds() error {
	LedsMap := GetLedsMap()
	for key := range *LedsMap {
		(*LedsMap)[key].controlLED((*LedsMap)[key].initialState)
		(*LedsMap)[key].startLedRoutine()
	}
	return nil
}

func MapLedsFromConfigs() error {
	// creates a leds map from the loaded configurations
	LedsMap := GetLedsMap()
	mutex.Lock()
	defer mutex.Unlock()
	// Map the existing LEDs
	// logger.Logger.Debugf("Config LED map: %v", *configs.GetConfigsLedMap())
	for _, ledInfo := range *configs.GetConfigsLedMap() {
		logger.Logger.Infof("Configure LED %s", ledInfo.Alias)
		(*LedsMap)[string(ledInfo.Alias)] = LED{
			path: ledInfo.OsPath,
			// pin:           gpio.NewPin(int(ledInfo.Bcm)),
			initialState:  ledInfo.InitialState,
			priorityMap:   make(map[PriorityMapKey]Priority),
			newPriorityCh: make(chan Priority, constants.CHANNEL_BUFFER_SIZE),
		}
	}
	return nil
}

func GetLedsMap() *map[string]LED {
	return &ledsMap
}

func GetLed(alias string) (*LED, error) {
	ledsMap := GetLedsMap()
	value, exists := (*ledsMap)[alias]
	if exists {
		return &value, nil
	} else {
		return &LED{}, fmt.Errorf("alias %s not found in ledsMap", alias)
	}
}

func (l LED) startLedRoutine() {
	go func() {
		priorityEndTime := time.Now().Add(constants.INFINITE_DURATION)
		for {
			select {
			case p := <-l.newPriorityCh:
				logger.Logger.Warnf("Setting LED to: %s", p.state)
				l.controlLED(p.state)
				priorityEndTime = p.priorityEndTime
			default:
				if time.Now().After(priorityEndTime) {
					l.manageLedState()
				}

				time.Sleep(time.Millisecond * 250)
			}

		}

	}()

}

func (l LED) controlLED(mode string) {
	switch mode {
	case string(hmi.On):
		l.onLED()
	case string(hmi.Off):
		l.offLED()
	case string(hmi.Blink):
		l.blinkLED()
	case string(hmi.Flick):
		l.flickLED()
	}
}

func (l LED) onLED() {
	os.WriteFile(fmt.Sprintf("%s/trigger", l.path), []byte("none"), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/brightness", l.path), []byte("1"), os.ModeExclusive)
}

func (l LED) offLED() {
	os.WriteFile(fmt.Sprintf("%s/trigger", l.path), []byte("none"), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/brightness", l.path), []byte("0"), os.ModeExclusive)
}

func (l LED) blinkLED() {
	os.WriteFile(fmt.Sprintf("%s/trigger", l.path), []byte("timer"), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/delay_off", l.path), []byte(viper.GetString(configs.ConfigTimersBlinkDown)), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/delay_on", l.path), []byte(viper.GetString(configs.ConfigTimersBlinkUp)), os.ModeExclusive)
}

func (l LED) flickLED() {
	os.WriteFile(fmt.Sprintf("%s/trigger", l.path), []byte("timer"), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/delay_off", l.path), []byte(viper.GetString(configs.ConfigTimersFlickDown)), os.ModeExclusive)
	os.WriteFile(fmt.Sprintf("%s/delay_on", l.path), []byte(viper.GetString(configs.ConfigTimersFlickUp)), os.ModeExclusive)

}

func (l LED) getPriorityMap() *map[PriorityMapKey]Priority {
	return &l.priorityMap
}

func (l LED) GetLedState() string {
	entry, err := l.findPrioritizedEntry()
	if err != nil {
		return l.initialState
	}
	return entry.state
}

func (l LED) findPrioritizedEntry() (Priority, error) {
	//find the most prioritize LED state listing in priorityMap

	var highestPriority PriorityMapKey
	highestPriorityLevel := uint(constants.LOWEST_PRIORITY)
	highestStateStrength := 0
	var stateStrength int
	priorityMap := l.getPriorityMap()

	if len(*priorityMap) != 0 {
		// find the strongest priority level (the lowest level exist in priorityMap with a valid end time)
		for priority_key := range *priorityMap {
			if priority_key.Level < highestPriorityLevel && time.Now().Before((*priorityMap)[priority_key].priorityEndTime) {
				highestPriorityLevel = priority_key.Level
			}
		}
		// find the priority listing with the strongest state according to the ledStateHierarchy
		for priority_key := range *priorityMap {
			if priority_key.Level == highestPriorityLevel && time.Now().Before((*priorityMap)[priority_key].priorityEndTime) {

				for index, state := range ledStateHierarchy {
					if string(state) == (*priorityMap)[priority_key].state {
						stateStrength = index
						break
					}
				}
				logger.Logger.Debugf("stateStrength: %d, highestStateStrength: %d", stateStrength, highestStateStrength)
				if stateStrength >= highestStateStrength {
					highestStateStrength = stateStrength
					highestPriority = priority_key
				}

			}
		}

	} else {
		return Priority{}, fmt.Errorf("priority map has no pending priorities")
	}

	logger.Logger.Debugf("PrioritizedEntry : %v", highestPriority)
	logger.Logger.Debugf("LED priority_level %d has these command IDs: %v", highestPriority.Level, l.getPriorityCommandIds(highestPriority.Level))

	return (*priorityMap)[highestPriority], nil

}
func (l LED) getPriorityCommandIds(priorityLevel uint) []string {
	var CommandIds []string
	priorityMap := l.getPriorityMap()
	for priority := range *priorityMap {
		if priority.Level == priorityLevel {
			CommandIds = append(CommandIds, priority.CommandId)
		}
	}
	return CommandIds
}

func (l LED) cleanTimedoutPriorities() error {
	// goes through the PriorityMap and removes the timed out listings
	priorityMap := l.getPriorityMap()
	for priority, p_struct := range *priorityMap {
		if time.Now().After(p_struct.priorityEndTime) {
			delete(*priorityMap, priority)
			logger.Logger.Warnf("Removing timed out priority %v from led priority map", priority)

		}
	}
	l.priorityMap = *priorityMap
	return nil
}
func (l LED) manageLedState() error {
	//Finds the right state to be set for the LED and send the priority entry to the newPriorityCh

	err := l.cleanTimedoutPriorities()
	if err != nil {
		logger.Logger.Error(err)
	}

	newPriority, err := l.findPrioritizedEntry()
	if err != nil {
		logger.Logger.Warnf("%s. Set to lowest priority entry", err)
		lowestPriorityEntry.state = l.initialState
		newPriority = lowestPriorityEntry
	}

	select {
	case l.newPriorityCh <- newPriority:
		logger.Logger.Warn("send new listing to channel.")
	default:
		return fmt.Errorf("Channel is not ready to receive, failed to send new listing to channel")
	}

	return nil
}

func calcLedPriorityEndTime(duration uint) (time.Time, error) {
	var priorityEndTime time.Time
	if duration == 0 {
		priorityEndTime = time.Now().Add(constants.INFINITE_DURATION)
	} else if duration > 0 {
		priorityEndTime = time.Now().Add(time.Duration(duration) * time.Second)
	} else {
		err := fmt.Errorf("Duration should be a positive number not: %d", duration)
		return time.Now(), err
	}
	return priorityEndTime, nil
}

func AddPriorityListing(ledType string, priorityKey PriorityMapKey, newState string, timeout uint, requestServiceUri string) error {
	ledsMap := GetLedsMap()
	mutex.Lock()
	defer mutex.Unlock()
	if led, exists := (*ledsMap)[ledType]; exists {

		logger.Logger.Debugf("key %s exists in LedsMap", ledType)
		logger.Logger.Debugf("priority is: %d", priorityKey.Level)
		if priorityKey.Level < constants.OVERRIDE_PRIORITY {
			err := fmt.Errorf("Priority is out of range: %d, value should be higher then %d", priorityKey.Level, constants.OVERRIDE_PRIORITY)
			logger.Logger.Error(err)
			return err
		}
		priorityMap := led.getPriorityMap()
		priorityEndTime, err := calcLedPriorityEndTime(timeout)
		formatDuration := priorityEndTime.Format("15:04:05")
		if timeout == 0 {
			formatDuration = "Infinite"
		}
		if err != nil {
			return err
		}

		if priority_listing, exists := (*priorityMap)[priorityKey]; exists {
			// priorityKey exists, update the listing with new values

			logger.Logger.Warnf("Priority key %v exists in LED priorityMap, updating it to state: %s, serviceUri: %s, priority end time: %s", priorityKey, newState, requestServiceUri, formatDuration)

			priority_listing.state = newState
			priority_listing.priorityEndTime = priorityEndTime
			priority_listing.serviceUri = requestServiceUri
			(*priorityMap)[priorityKey] = priority_listing

		} else {
			// the priorityKey doesn't exist adding it to priorityMap
			logger.Logger.Warnf("Priority key %v doesn't found in LED %s priorityMap, creating new priority listing", priorityKey, ledType)
			entry := Priority{
				state:           newState,
				priorityEndTime: priorityEndTime,
				serviceUri:      requestServiceUri,
			}
			logger.Logger.Warnf("Updating LED %s with new priority listing, key:%v, value: Priority{%s, %s}", ledType, priorityKey, newState, formatDuration)
			(*priorityMap)[priorityKey] = entry
		}

		(*ledsMap)[ledType] = led
		(*ledsMap)[ledType].manageLedState()
	}
	return nil
}

func RemovePriorityListing(ledType string, priorityKey PriorityMapKey) error {
	ledsMap := GetLedsMap()
	mutex.Lock()
	defer mutex.Unlock()
	if led, exists := (*ledsMap)[ledType]; exists {

		logger.Logger.Debugf("key %s exists in LedsMap", ledType)
		logger.Logger.Debugf("priority is: %d", priorityKey.Level)
		priorityMap := led.getPriorityMap()

		if _, exists := (*priorityMap)[priorityKey]; exists {
			// release the priority listing
			logger.Logger.Warnf("Removes priorityKey %v from LED %s priorityMap", priorityKey, ledType)
			delete(*priorityMap, priorityKey)
			logger.Logger.Debugf("LED %s priority_level %d has these command IDs: %v", ledType, priorityKey.Level, led.getPriorityCommandIds(priorityKey.Level))
		} else {
			logger.Logger.Warnf("Priority key %v doesn't exists in LED %s, nothing to release", priorityKey, ledType)
			return nil
		}
		(*ledsMap)[ledType] = led
		(*ledsMap)[ledType].manageLedState()
	}
	return nil
}

func OverrideLedState(alias string, overrideState string, overrideTimeout uint) {
	// Service override the LED priority
	ledsMap := GetLedsMap()
	if led, exists := (*ledsMap)[alias]; exists {
		priorityKey := PriorityMapKey{
			Level:     constants.OVERRIDE_PRIORITY,
			CommandId: constants.GetServiceName(),
		}
		AddPriorityListing(alias, priorityKey, overrideState, overrideTimeout, priorityKey.CommandId)
		(*ledsMap)[alias] = led
	} else {
		logger.Logger.Errorf("%s dosen't exist in LEDs Map", alias)
	}
}

func DisableOverrideLedState(alias string) {
	// Disable the service LED priority override
	ledsMap := GetLedsMap()
	if led, exists := (*ledsMap)[alias]; exists {
		priorityKey := PriorityMapKey{
			Level:     constants.OVERRIDE_PRIORITY,
			CommandId: constants.GetServiceName(),
		}
		AddPriorityListing(alias, priorityKey, string(hmi.Off), constants.INFINITE_TIMEOUT, priorityKey.CommandId)
		(*ledsMap)[alias] = led
	} else {
		logger.Logger.Errorf("%s dosen't exist in LEDs Map", alias)
	}
}

func DisableAllOverride() {
	ledsMap := GetLedsMap()
	for key := range *ledsMap {
		DisableOverrideLedState(key)
	}
}

func OverrideAllLeds(state string, overrideTimeout uint) {
	ledsMap := GetLedsMap()
	for key := range *ledsMap {
		OverrideLedState(key, state, overrideTimeout)
	}
}

func clearLedPriorityMap(ledType string) {
	ledsMap := GetLedsMap()
	if led, exists := (*ledsMap)[ledType]; exists {
		logger.Logger.Warnf("Clearing LED: %s, priority map", ledType)
		led.priorityMap = make(map[PriorityMapKey]Priority)
		(*ledsMap)[ledType] = led
	}
}

func clearAllLedsPriorityMaps() {
	logger.Logger.Warn("Clearing all LEDs priority maps")
	ledsMap := GetLedsMap()
	for key := range *ledsMap {
		clearLedPriorityMap(key)
	}
}

func ClearServiceUriCommands(serviceUri string) error {
	logger.Logger.Warnf("Clearing all service %s commands from priority maps", serviceUri)
	ledsMap := GetLedsMap()
	for ledType, led := range *ledsMap {
		logger.Logger.Debugf("Clearing service commands from LED: %s priority map", ledType)
		for priorityMapKey, priority := range led.priorityMap {
			if priority.serviceUri == serviceUri {
				delete(led.priorityMap, priorityMapKey)
			}
		}
		logger.Logger.Debugf("LED %s, priorityMAP left with these priorities:", ledType)
		for priorityMapKey, priority := range led.priorityMap {
			logger.Logger.Debugf("priorityMapKey: %v, priority: %v", priorityMapKey, priority)
		}
		(*ledsMap)[ledType] = led
		led.manageLedState()
	}
	return nil
}
