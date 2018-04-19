package epoch

import (
	"time"
)

func nHourly(e Epoch, n int, prev time.Time, next time.Time) bool {
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= float64(n) {
		return true
	}
	return prev.Hour()/n != next.Hour()/n
}

type EightHourly struct{}

func (EightHourly) GetData() Data {
	return Data{
		"Once every eight hours",
		"The last PR merge commit of eight-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:08:00, etc..",
		time.Hour * 8,
		time.Hour * 8,
	}
}

func (e EightHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 8, prev, next)
}

type FourHourly struct{}

func (FourHourly) GetData() Data {
	return Data{
		"Once every four hours",
		"The last PR merge commit of four-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:04:00, etc..",
		time.Hour * 4,
		time.Hour * 4,
	}
}

func (e FourHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 4, prev, next)
}

type TwoHourly struct{}

func (TwoHourly) GetData() Data {
	return Data{
		"Once every two hours",
		"The last PR merge commit of two-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:02:00, etc..",
		time.Hour * 2,
		time.Hour * 2,
	}
}

func (e TwoHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 2, prev, next)
}
