package epoch

import (
	"time"
)

var halfDaily = &Epoch{
	MinDuration: time.Hour * 12,
	MaxDuration: time.Hour * 12,
	IsEpochal:   halfDailyIsEpochal,
}

var quarterDaily = &Epoch{
	MinDuration: time.Hour * 6,
	MaxDuration: time.Hour * 6,
	IsEpochal:   quarterDailyIsEpochal,
}

var eighthDaily = &Epoch{
	MinDuration: time.Hour * 3,
	MaxDuration: time.Hour * 3,
	IsEpochal:   eighthDailyIsEpochal,
}

func halfDailyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return halfDailyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 12 {
		return true
	}
	return prev.Hour()/12 != next.Hour()/12
}

func quarterDailyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return quarterDailyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 6 {
		return true
	}
	return prev.Hour()/6 != next.Hour()/6
}

func eighthDailyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return eighthDailyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 3 {
		return true
	}
	return prev.Hour()/3 != next.Hour()/3
}

// GetHalfDaily generates an Epoch that changes at the beginning of every day, and half way through every day.
func GetHalfDaily() *Epoch {
	return halfDaily
}

// GetQuarterDaily generates an Epoch that changes at the beginning of every day, every six hours thereafter.
func GetQuarterDaily() *Epoch {
	return quarterDaily
}

// GetEighthDaily generates an Epoch that changes at the beginning of every day, every six hours thereafter.
func GetEighthDaily() *Epoch {
	return eighthDaily
}

var announcerEpochs = []*Epoch{weekly, daily, halfDaily, quarterDaily, eighthDaily, hourly}

// GetAnnouncerEpochs generates a []*Epoch in descending order of epoch length, where each Epoch corresponds to a an epoch managed by the WPT revision announcer.
func GetAnnouncerEpochs() []*Epoch {
	return announcerEpochs
}
