package settings

import "github.com/Jhnvlglmlbrt/monitoring-certs/data"

type accountSettings struct {
	MaxTrackings     int
	Webhooks         bool
	SlackIntegration bool
	TeamsIntegration bool
}

var Account = map[data.Plan]accountSettings{
	data.PlanStarter: {
		MaxTrackings: 2,
	},
	data.PlanBusiness: {
		MaxTrackings:     20,
		Webhooks:         true,
		SlackIntegration: true,
	},
	data.PlanEnterprise: {
		MaxTrackings:     200,
		Webhooks:         true,
		SlackIntegration: true,
		TeamsIntegration: true,
	},
}
