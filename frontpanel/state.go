package frontpanel

import (
	"encoding/json"
)

func (a *App) updateState() {
	jsData, err := json.Marshal(a.configState)
	if err != nil {
		a.logger.Info().Msgf("failed to marshal json data: %v", err)
		return
	}

	err = a.NDKAgent.UpdateState(AppRoot, string(jsData))
	if err != nil {
		a.logger.Error().Msgf("failed to update state: %v", err)
	}
}
