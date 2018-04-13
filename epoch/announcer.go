package epoch

import (
	"time"
)

var halfDaily *Epoch
var quarterDaily *Epoch
var eighthDaily *Epoch

func init() {
	halfDaily = &Epoch{
		MinDuration: time.Hour * 12,
		MaxDuration: time.Hour * 12,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 12 {
				return true
			}
			return prev.Hour() < 12 && next.Hour() >= 12
		},
	}
	quarterDaily = &Epoch{
		MinDuration: time.Hour * 6,
		MaxDuration: time.Hour * 6,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 6 {
				return true
			}
			return prev.Hour()/4 != next.Hour()/4
		},
	}
	eighthDaily = &Epoch{
		MinDuration: time.Hour * 3,
		MaxDuration: time.Hour * 3,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 3 {
				return true
			}
			return prev.Hour()/8 != next.Hour()/8
		},
	}
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

var announcerEpochs []*Epoch

func init() {
	announcerEpochs = []*Epoch{GetWeekly(), GetDaily(), GetHalfDaily(), GetQuarterDaily(), GetEighthDaily(), GetHourly()}
}

// GetAnnouncerEpochs generates a []*Epoch in descending order of epoch length, where each Epoch corresponds to a an epoch managed by the WPT revision announcer.
func GetAnnouncerEpochs() []*Epoch {
	return announcerEpochs
}
