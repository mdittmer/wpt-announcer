package epoch_test

import (
	"testing"
	"time"

	"github.com/mdittmer/wpt-announcer/epoch"
	"github.com/stretchr/testify/assert"
)

var monthly *epoch.Epoch
var weekly *epoch.Epoch
var daily *epoch.Epoch
var hourly *epoch.Epoch
var nilBasis epoch.Basis

func init() {
	monthly = epoch.GetMonthly()
	weekly = epoch.GetWeekly()
	daily = epoch.GetDaily()
	hourly = epoch.GetHourly()
	nilBasis = epoch.Basis{}
}

//
// Monthly
//

func TestIsMonthly_Close(t *testing.T) {
	justPrior := time.Date(2018, 3, 31, 23, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, monthly.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, monthly.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsMonthly_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, monthly.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, monthly.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotMonthly_Close(t *testing.T) {
	lastInstant := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 5, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, monthly.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, monthly.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotMonthly_Far(t *testing.T) {
	monthStart := time.Date(2018, 3, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := time.Date(2018, 3, 31, 23, 59, 59, 999999999, time.UTC)
	assert.False(t, monthly.IsEpochal(&monthStart, &monthEnd, &nilBasis))
	assert.False(t, monthly.IsEpochal(&monthEnd, &monthStart, &nilBasis))
}

//
// Weekly
//

func TestIsWeekly_Close(t *testing.T) {
	justPrior := time.Date(2018, 3, 31, 23, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, weekly.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, weekly.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsWeekly_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, weekly.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, weekly.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotWeekly_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, weekly.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, weekly.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotWeekly_Far(t *testing.T) {
	weekStart := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2018, 4, 7, 23, 59, 59, 999999999, time.UTC)
	assert.False(t, weekly.IsEpochal(&weekStart, &weekEnd, &nilBasis))
	assert.False(t, weekly.IsEpochal(&weekEnd, &weekStart, &nilBasis))
}

//
// Daily
//

func TestIsDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC)
	assert.True(t, daily.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, daily.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, daily.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, daily.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, daily.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, daily.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC)
	assert.False(t, daily.IsEpochal(&dayStart, &dayEnd, &nilBasis))
	assert.False(t, daily.IsEpochal(&dayEnd, &dayStart, &nilBasis))
}

//
// Hourly
//

func TestIsHourly_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 0, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 2, 1, 0, 0, 0, time.UTC)
	assert.True(t, hourly.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, hourly.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsHourly_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, hourly.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, hourly.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotHourly_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, hourly.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, hourly.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotHourly_Far(t *testing.T) {
	hourStart := time.Date(2018, 4, 1, 1, 0, 0, 0, time.UTC)
	hourEnd := time.Date(2018, 4, 1, 1, 59, 59, 999999999, time.UTC)
	assert.False(t, hourly.IsEpochal(&hourStart, &hourEnd, &nilBasis))
	assert.False(t, hourly.IsEpochal(&hourEnd, &hourStart, &nilBasis))
}

//
// GetGregorianEpochs
//

func TestGregorian(t *testing.T) {
	gregorian := epoch.GetGregorianEpochs()
	assert.True(t, len(gregorian) == 4)
	assert.True(t, gregorian[0] == monthly)
	assert.True(t, gregorian[1] == weekly)
	assert.True(t, gregorian[2] == daily)
	assert.True(t, gregorian[3] == hourly)
}
