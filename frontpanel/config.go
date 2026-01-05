package frontpanel

import (
	"encoding/json"
)

type ConfigState struct {
	// URL links to a high-resolution image of the front panel
	URL string `json:"url,omitempty"`
}

func (a *App) loadConfig() {
	a.configState = &ConfigState{} // clear the configState
	if a.NDKAgent.Notifications.FullConfig != nil {
		err := json.Unmarshal(a.NDKAgent.Notifications.FullConfig, a.configState)
		if err != nil {
			a.logger.Error().Err(err).Msg("Failed to unmarshal config")
		}
	}
}

func (a *App) processConfig() {
	chassisType, err := a.getChassisType()
	if err != nil {
		a.logger.Error().Msgf("failed to get chassis type: %v", err)
		return
	}

	chassisUrl := "not supported"
	if chassisDef, ok := chassisImages[chassisType]; ok {
		chassisUrl = chassisDef.URL
	}

	a.configState.URL = chassisUrl
}
