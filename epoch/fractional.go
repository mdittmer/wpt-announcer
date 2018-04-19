package epoch

import (
	"time"
)

func fixedData(t time.Duration) Data {
	return Data{t, t}
}

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
	return fixedData(time.Hour * 8)
}

func (e EightHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 8, prev, next)
}

type FourHourly struct{}

func (FourHourly) GetData() Data {
	return fixedData(time.Hour * 4)
}

func (e FourHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 4, prev, next)
}

type TwoHourly struct{}

func (TwoHourly) GetData() Data {
	return fixedData(time.Hour * 2)
}

func (e TwoHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 2, prev, next)
}
