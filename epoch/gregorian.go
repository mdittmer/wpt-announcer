package epoch

import (
	"time"
)

type Monthly struct{}

func (Monthly) GetData() Data {
	return Data{
		time.Hour * 24 * 28,
		time.Hour * 24 * 31,
	}
}

func (Monthly) IsEpochal(prev time.Time, next time.Time) bool {
	if prev.Year() != next.Year() {
		return true
	}
	return prev.Month() != next.Month()
}

type Weekly struct{}

func (Weekly) GetData() Data {
	return Data{
		time.Hour * 24 * 7,
		time.Hour * 24 * 7,
	}
}

func (e Weekly) IsEpochal(prev time.Time, next time.Time) bool {
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 24*7 {
		return true
	}
	return prev.Weekday() > next.Weekday()
}

type Daily struct{}

func (Daily) GetData() Data {
	return Data{
		time.Hour * 24,
		time.Hour * 24,
	}
}

func (e Daily) IsEpochal(prev time.Time, next time.Time) bool {
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 24 {
		return true
	}
	return prev.Day() != next.Day()
}

type Hourly struct{}

func (Hourly) GetData() Data {
	return Data{
		time.Hour,
		time.Hour,
	}
}

func (e Hourly) IsEpochal(prev time.Time, next time.Time) bool {
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 1 {
		return true
	}
	return prev.Hour() != next.Hour()
}
