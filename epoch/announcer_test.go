package epoch_test

import (
	"testing"
	"time"

	"github.com/mdittmer/wpt-announcer/epoch"
	"github.com/stretchr/testify/assert"
)

var halfDaily *epoch.Epoch
var quarterDaily *epoch.Epoch
var eighthDaily *epoch.Epoch

func init() {
	halfDaily = epoch.GetHalfDaily()
	quarterDaily = epoch.GetQuarterDaily()
	eighthDaily = epoch.GetEighthDaily()
}

//
// HalfDaily
//

func TestIsHalfDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 11, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 12, 0, 0, 0, time.UTC)
	assert.True(t, halfDaily.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, halfDaily.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsHalfDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, halfDaily.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, halfDaily.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotHalfDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, halfDaily.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, halfDaily.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotHalfDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 12, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC)
	assert.False(t, halfDaily.IsEpochal(&dayStart, &dayEnd, &nilBasis))
	assert.False(t, halfDaily.IsEpochal(&dayEnd, &dayStart, &nilBasis))
}

//
// QuarterDaily
//

func TestIsQuarterDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 5, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 6, 0, 0, 0, time.UTC)
	assert.True(t, quarterDaily.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, quarterDaily.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsQuarterDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, quarterDaily.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, quarterDaily.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotQuarterDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, quarterDaily.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, quarterDaily.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotQuarterDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 6, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 11, 59, 59, 999999999, time.UTC)
	assert.False(t, quarterDaily.IsEpochal(&dayStart, &dayEnd, &nilBasis))
	assert.False(t, quarterDaily.IsEpochal(&dayEnd, &dayStart, &nilBasis))
}

//
// EighthDaily
//

func TestIsEighthDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 2, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 3, 0, 0, 0, time.UTC)
	assert.True(t, eighthDaily.IsEpochal(&justPrior, &justAfter, &nilBasis))
	assert.True(t, eighthDaily.IsEpochal(&justAfter, &justPrior, &nilBasis))
}

func TestIsEighthDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, eighthDaily.IsEpochal(&lastYear, &thisYear, &nilBasis))
	assert.True(t, eighthDaily.IsEpochal(&thisYear, &lastYear, &nilBasis))
}

func TestIsNotEighthDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, eighthDaily.IsEpochal(&lastInstant, &nextInstant, &nilBasis))
	assert.False(t, eighthDaily.IsEpochal(&nextInstant, &lastInstant, &nilBasis))
}

func TestIsNotEighthDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 3, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 5, 59, 59, 999999999, time.UTC)
	assert.False(t, eighthDaily.IsEpochal(&dayStart, &dayEnd, &nilBasis))
	assert.False(t, eighthDaily.IsEpochal(&dayEnd, &dayStart, &nilBasis))
}
