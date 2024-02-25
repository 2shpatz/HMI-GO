package apps_monitor

import (
	"eos/hmi-service/pkg/health_monitor/balena"
	"eos/hmi-service/pkg/hmi/leds"
	"eos/hmi-service/pkg/utils/configs"
	"eos/hmi-service/pkg/utils/logger"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/services/hmi"
)

const (
	POLL_INTERVAL = 5
)

type DeviceStatus string

const (
	STOPPING    DeviceStatus = "Stopping"
	STARTING                 = "Starting"
	DOWNLOADING              = "Downloading"
	RUNNING                  = "running"
	INSTALLING               = "Installing"
	IDLE                     = "Idle"
)

type Monitor struct {
	BalenaSdk balena.Balena
}

func Run() {

	// TBR after controlling the power led
	priorityKey := leds.PriorityMapKey{
		Level:     10,
		CommandId: "running",
	}
	err := leds.AddPriorityListing(string(hmi.Power), priorityKey, string(hmi.On), 0, "apps-monitor")
	if err != nil {
		logger.Logger.Errorf("Couldn't set temporary power LED state: %s", err)
	}
	////////////////////////////////////////

	balena := balena.Balena{
		HttpClient:   &http.Client{},
		BalenaApi:    viper.GetString(configs.ConfigServiceSupervisorAddress),
		BalenaApiKey: viper.GetString(configs.ConfigServiceSupervisorApiKey),
		Services:     make(map[string]balena.BalenaService),
	}
	appsMonitor := Monitor{
		BalenaSdk: balena,
	}

	appsMonitor.StartPollSupervisor()
}

func (am *Monitor) StartPollSupervisor() {
	err := am.BalenaSdk.CheckApiKey()
	if err != nil {
		return
	}
	logger.Logger.Debug("Start polling supervisor")
	go func() {

		for {
			err := am.BalenaSdk.GetDevice()
			if err != nil {
				logger.Logger.Errorf("getDevice error: %s", err)
			} else if am.BalenaSdk.BalenaDevice.UpdatePending {
				logger.Logger.Warnf("System is updating...")
				logger.Logger.Debugf("Device status: %s\nUpdateDownloaded?: %t\nUpdatePending?: %t", am.BalenaSdk.BalenaDevice.Status, am.BalenaSdk.BalenaDevice.UpdateDownloaded, am.BalenaSdk.BalenaDevice.UpdatePending)
				leds.OverrideLedState(string(hmi.Power), string(hmi.Flick), POLL_INTERVAL+1)
			}

			// err = am.BalenaSdk.GetServices()
			// if err != nil {
			// 	logger.Logger.Errorf("getServices error: %s", err)
			// } else if len(am.BalenaSdk.Services) > 0 {
			// 	for serviceName, serviceData := range am.BalenaSdk.Services {
			// 		logger.Logger.Debugf("Service: %s, status is: %v", serviceName, serviceData)
			// 		// if serviceData.Status == DOWNLOADING || serviceData.Status == INSTALLING {

			// 			logger.Logger.Warnf("Service: %s, is updating with status: %s, download progress: %d", serviceName, serviceData.Status, serviceData.DownloadProgress)
			// 			leds.OverrideLedState(string(hmi.Power), string(hmi.Flick), POLL_INTERVAL+1)
			// 			break
			// 		}
			// 	}
			// } else {
			// 	logger.Logger.Error("No Service was found")
			// }
			time.Sleep(POLL_INTERVAL * time.Second)
		}
	}()
}
