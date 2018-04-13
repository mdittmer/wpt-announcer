package epoch

import (
	"time"
)

var monthly *Epoch
var weekly *Epoch
var daily *Epoch
var hourly *Epoch

func monthlyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.Year() != next.Year() {
		return true
	}
	return prev.Month() != next.Month()
}

func weeklyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return weeklyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 24*7 {
		return true
	}
	return prev.Weekday() > next.Weekday()
}

func dailyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return dailyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 24 {
		return true
	}
	return prev.Day() != next.Day()
}

func hourlyIsEpochal(prev *time.Time, next *time.Time, basis *Basis) bool {
	if prev.After(*next) {
		return hourlyIsEpochal(next, prev, basis)
	}
	if next.Sub(*prev).Hours() >= 1 {
		return true
	}
	return prev.Hour() != next.Hour()
}

func init() {
	monthly = &Epoch{
		MinDuration: time.Hour * 24 * 28,
		MaxDuration: time.Hour * 24 * 31,
		IsEpochal:   monthlyIsEpochal,
	}
	weekly = &Epoch{
		MinDuration: time.Hour * 24 * 7,
		MaxDuration: time.Hour * 24 * 7,
		IsEpochal:   weeklyIsEpochal,
	}
	daily = &Epoch{
		MinDuration: time.Hour * 24,
		MaxDuration: time.Hour * 24,
		IsEpochal:   dailyIsEpochal,
	}
	hourly = &Epoch{
		MinDuration: time.Hour,
		MaxDuration: time.Hour,
		IsEpochal:   hourlyIsEpochal,
	}
}

// GetMonthly generates an Epoch that changes at the beginning of every month according to time.Time.Month() enumeration.
func GetMonthly() *Epoch {
	return monthly
}

// GetWeekly generates an Epoch that changes at the beginning of every week according to time.Time.Weekday() enumeration.
func GetWeekly() *Epoch {
	return weekly
}

// GetDaily generates an Epoch that changes at the beginning of every day according to time.Time.Day() enumeration.
func GetDaily() *Epoch {
	return daily
}

// GetHourly generates an Epoch that changes at the beginning of every hour of the day according to time.Time.Hour() enumeration.
func GetHourly() *Epoch {
	return hourly
}

var gregorianEpochs []*Epoch

func init() {
	gregorianEpochs = []*Epoch{GetMonthly(), GetWeekly(), GetDaily(), GetHourly()}
}

// GetGregorianEpochs generates a []Epoch in descending order of epoch length, where each Epoch corresponds to a Gregorian calendar measure of time.
func GetGregorianEpochs() []*Epoch {
	return gregorianEpochs
}
