package frontpanel

import (
	"context"
	"strings"

	"github.com/openconfig/gnmic/pkg/api"
	"github.com/rs/zerolog"
	"github.com/srl-labs/bond"
)

const (
	AppName = "front-panel"
	AppRoot = "/platform/" + AppName
)

type App struct {
	Name string
	// configState holds the application configuration and state.
	configState *ConfigState
	logger      *zerolog.Logger
	NDKAgent    *bond.Agent
}

func New(logger *zerolog.Logger, agent *bond.Agent) *App {
	return &App{
		Name:     AppName,
		NDKAgent: agent,

		logger: logger,
	}
}

func (a *App) Start(ctx context.Context) {
	for {
		select {
		case <-a.NDKAgent.Notifications.FullConfigReceived:
			a.logger.Debug().Msg("Received full config")

			a.loadConfig()

			a.processConfig()

			a.updateState()

		case <-ctx.Done():
			return
		}
	}
}

func (a *App) getChassisType() (string, error) {
	a.logger.Debug().Msg("Fetching chassis type")

	// Fetch admin states first
	getReqType, err := bond.NewGetRequest("/platform/chassis/type", api.EncodingPROTO())
	if err != nil {
		return "", err
	}

	getRespType, err := a.NDKAgent.GetWithGNMI(getReqType)
	if err != nil {
		return "", err
	}

	a.logger.Debug().Msgf("GetResponse Type: %+v", getRespType)

	chassisType := getRespType.GetNotification()[0].GetUpdate()[0].Val.GetStringVal()

	return chassisType, nil
}

func (a *App) getFrontPorts() (*map[string]string, error) {
	frontPortStates := make(map[string]string)

	a.logger.Debug().Msg("Fetching front port states")

	// Fetch admin states first
	getReqAdmin, err := bond.NewGetRequest("/interface[name=ethernet-*]/admin-state", api.EncodingPROTO())
	if err != nil {
		return nil, err
	}

	getRespAdmin, err := a.NDKAgent.GetWithGNMI(getReqAdmin)
	if err != nil {
		return nil, err
	}

	a.logger.Debug().Msgf("GetResponse AdminState: %+v", getRespAdmin)

	// Populate front port states using admin state as baseline.
	for _, notification := range getRespAdmin.GetNotification() {
		for _, update := range notification.GetUpdate() {
			ifName := ""
			for _, elem := range update.Path.Elem {
				if elem.Name == "interface" {
					ifName = elem.Key["name"]
				}
			}

			if ifName == "" {
				continue
			}

			adminState := strings.ToLower(strings.TrimSpace(update.Val.GetStringVal()))
			if adminState == "disable" || adminState == "disabled" || adminState == "down" {
				frontPortStates[ifName] = "admin-down"
				continue
			}

			frontPortStates[ifName] = "admin-up-oper-down"
		}
	}

	// Fetch oper states
	getReqOper, err := bond.NewGetRequest("/interface[name=ethernet-*]/oper-state", api.EncodingPROTO())
	if err != nil {
		return nil, err
	}

	getRespOper, err := a.NDKAgent.GetWithGNMI(getReqOper)
	if err != nil {
		return nil, err
	}

	a.logger.Debug().Msgf("GetResponse OperState: %+v", getRespOper)

	// For admin-up ports, use the current oper state.
	for _, notification := range getRespOper.GetNotification() {
		for _, update := range notification.GetUpdate() {
			ifName := ""
			for _, elem := range update.Path.Elem {
				if elem.Name == "interface" {
					ifName = elem.Key["name"]
				}
			}

			// Skip interfaces that are in admin-down state.
			if ifName == "" || frontPortStates[ifName] == "admin-down" {
				continue
			}

			operState := strings.ToLower(strings.TrimSpace(update.Val.GetStringVal()))
			if operState == "up" || operState == "oper-up" {
				frontPortStates[ifName] = "admin-up-oper-up"
				continue
			}

			frontPortStates[ifName] = "admin-up-oper-down"
		}
	}

	a.logger.Debug().Msgf("Front port states: %+v", frontPortStates)

	return &frontPortStates, nil
}
