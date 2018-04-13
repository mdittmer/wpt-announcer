package epoch_test

import (
	"testing"
	"time"

	"github.com/mdittmer/wpt-announcer/epoch"
	"github.com/stretchr/testify/assert"
)

var monthly = epoch.GetMonthly()
var weekly = epoch.GetWeekly()
var nilBasis = epoch.Basis{}

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
