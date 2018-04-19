package main

import (
	"github.com/mdittmer/wpt-announcer/api"
	"github.com/mdittmer/wpt-announcer/epoch"
)

var epochs = []epoch.Epoch{
	epoch.Weekly{},
	epoch.Daily{},
	epoch.EightHourly{},
	epoch.FourHourly{},
	epoch.TwoHourly{},
	epoch.Hourly{},
}

var apiEpochs = make([]api.Epoch, 0)

var latestGetRevisions = make(map[epoch.Epoch]int)

func init() {
	for _, e := range epochs {
		apiEpochs = append(apiEpochs, api.FromEpoch(e))
		latestGetRevisions[e] = 1
	}
}
