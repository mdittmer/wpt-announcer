package epoch

import (
	"time"
)

// GetMonthly generates an Epoch that changes at the beginning of every month according to time.Time.Month() enumeration.
func GetMonthly() Epoch {
	return Epoch{
		MinDuration: time.Hour * 24 * 28,
		MaxDuration: time.Hour * 24 * 31,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if prev.Year() != next.Year() {
				return true
			}
			return prev.Month() != next.Month()
		},
	}
}

// GetWeekly generates an Epoch that changes at the beginning of every week according to time.Time.Weekday() enumeration.
func GetWeekly() Epoch {
	return Epoch{
		MinDuration: time.Hour * 24 * 7,
		MaxDuration: time.Hour * 24 * 7,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 24*7 {
				return true
			}
			return prev.Weekday() > next.Weekday()
		},
	}
}

// GetDaily generates an Epoch that changes at the beginning of every day according to time.Time.Day() enumeration.
func GetDaily() Epoch {
	return Epoch{
		MinDuration: time.Hour * 24,
		MaxDuration: time.Hour * 24,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 24 {
				return true
			}
			return prev.Day() != next.Day()
		},
	}
}

// GetHourly generates an Epoch that changes at the beginning of every hour of the day according to time.Time.Hour() enumeration.
func GetHourly() Epoch {
	return Epoch{
		MinDuration: time.Hour,
		MaxDuration: time.Hour,
		IsEpochal: func(prev *time.Time, next *time.Time, basis *Basis) bool {
			if next.Sub(*prev).Hours() >= 1 {
				return true
			}
			return prev.Hour() != next.Hour()
		},
	}
}

// GetGregorianEpochs generates a []Epoch in descending order of epoch length, where each Epoch corresponds to a Gregorian calendar measure of time.
func GetGregorianEpochs() []Epoch {
	return []Epoch{GetMonthly(), GetWeekly(), GetDaily(), GetHourly()}
}
