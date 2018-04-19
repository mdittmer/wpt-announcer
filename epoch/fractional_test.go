package epoch_test

import (
	"testing"
	"time"

	"github.com/mdittmer/wpt-announcer/epoch"
	"github.com/stretchr/testify/assert"
)

var eightHourly = epoch.EightHourly{}
var fourHourly = epoch.FourHourly{}
var twoHourly = epoch.TwoHourly{}

func testClosePositive(t *testing.T, e epoch.Epoch) {
	n := int(e.GetData().MaxDuration.Hours())
	justPrior := time.Date(2018, 4, 1, n-1, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, n, 0, 0, 0, time.UTC)
	assert.True(t, e.IsEpochal(justPrior, justAfter))
	assert.True(t, e.IsEpochal(justAfter, justPrior))
}

var lastYear = time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
var thisYear = time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)

func testFarPositive(t *testing.T, e epoch.Epoch) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, e.IsEpochal(lastYear, thisYear))
	assert.True(t, e.IsEpochal(thisYear, lastYear))
}

var startDay = time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
var justAfterStartDay = time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)

func testCloseNegative(t *testing.T, e epoch.Epoch) {
	assert.False(t, e.IsEpochal(startDay, justAfterStartDay))
	assert.False(t, e.IsEpochal(justAfterStartDay, startDay))
}

func testFarNegative(t *testing.T, e epoch.Epoch) {
	n := int(e.GetData().MaxDuration.Hours())
	start := time.Date(2018, 4, 1, n, 0, 0, 0, time.UTC)
	justBeforeEnd := time.Date(2018, 4, 1, 2*n-1, 59, 59, 999999999, time.UTC)
	assert.False(t, e.IsEpochal(start, justBeforeEnd))
	assert.False(t, e.IsEpochal(justBeforeEnd, start))
}

//
// EightHourly
//

func TestIsEightHourly_Close(t *testing.T) {
	testClosePositive(t, eightHourly)
}

func TestIsEightHourly_Far(t *testing.T) {
	testFarPositive(t, eightHourly)
}

func TestIsNotEightHourly_Close(t *testing.T) {
	testCloseNegative(t, eightHourly)
}

func TestIsNotEightHourly_Far(t *testing.T) {
	testFarNegative(t, eightHourly)
}

//
// QuarterDaily
//
/*
func TestIsQuarterDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 5, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 6, 0, 0, 0, time.UTC)
	assert.True(t, quarterDaily.IsEpochal(&justPrior, &justAfter))
	assert.True(t, quarterDaily.IsEpochal(&justAfter, &justPrior))
}

func TestIsQuarterDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, quarterDaily.IsEpochal(&lastYear, &thisYear))
	assert.True(t, quarterDaily.IsEpochal(&thisYear, &lastYear))
}

func TestIsNotQuarterDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, quarterDaily.IsEpochal(&lastInstant, &nextInstant))
	assert.False(t, quarterDaily.IsEpochal(&nextInstant, &lastInstant))
}

func TestIsNotQuarterDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 6, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 11, 59, 59, 999999999, time.UTC)
	assert.False(t, quarterDaily.IsEpochal(&dayStart, &dayEnd))
	assert.False(t, quarterDaily.IsEpochal(&dayEnd, &dayStart))
}

//
// EighthDaily
//

func TestIsEighthDaily_Close(t *testing.T) {
	justPrior := time.Date(2018, 4, 1, 2, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 3, 0, 0, 0, time.UTC)
	assert.True(t, eighthDaily.IsEpochal(&justPrior, &justAfter))
	assert.True(t, eighthDaily.IsEpochal(&justAfter, &justPrior))
}

func TestIsEighthDaily_Far(t *testing.T) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, eighthDaily.IsEpochal(&lastYear, &thisYear))
	assert.True(t, eighthDaily.IsEpochal(&thisYear, &lastYear))
}

func TestIsNotEighthDaily_Close(t *testing.T) {
	lastInstant := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	nextInstant := time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
	assert.False(t, eighthDaily.IsEpochal(&lastInstant, &nextInstant))
	assert.False(t, eighthDaily.IsEpochal(&nextInstant, &lastInstant))
}

func TestIsNotEighthDaily_Far(t *testing.T) {
	dayStart := time.Date(2018, 4, 1, 3, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2018, 4, 1, 5, 59, 59, 999999999, time.UTC)
	assert.False(t, eighthDaily.IsEpochal(&dayStart, &dayEnd))
	assert.False(t, eighthDaily.IsEpochal(&dayEnd, &dayStart))
}

//
// Relationships between epochs
//

func TestEpochAlignment(t *testing.T) {
	epochs := epoch.GetAnnouncerEpochs()
	justPrior := time.Date(2018, 3, 31, 23, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
	for _, epoch := range epochs {
		assert.True(t, epoch.IsEpochal(&justPrior, &justAfter))
		assert.True(t, epoch.IsEpochal(&justAfter, &justPrior))
	}
}
*/
