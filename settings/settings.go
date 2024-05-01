package settings

import "github.com/Jhnvlglmlbrt/monitoring-certs/data"

type accountSettings struct {
	MaxTrackings       int
	Webhooks           bool
	DiscordIntegration bool
	TeamsIntegration   bool
}

var Account = map[data.Plan]accountSettings{
	data.PlanStarter: {
		MaxTrackings: 2,
	},
	data.PlanBusiness: {
		MaxTrackings:       50,
		Webhooks:           true,
		DiscordIntegration: true,
	},
	data.PlanEnterprise: {
		MaxTrackings:       200,
		Webhooks:           true,
		DiscordIntegration: true,
		TeamsIntegration:   true,
	},
}
