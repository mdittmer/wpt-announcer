package epoch

import "time"

// Basis is used by EpochalPredicate functions that require a basis for calculating the beginning of each epoch.
type Basis struct {
	Start *time.Time
	Now   *time.Time
}

// Epoch encapsulates a pattern in time during which new epochs begin at regular intervals.
type Epoch struct {
	MinDuration time.Duration
	MaxDuration time.Duration
	IsEpochal   EpochalPredicate
}

// EpochalPredicate is a predicate that determines whether a new epoch begins between prev and next.
type EpochalPredicate func(prev *time.Time, next *time.Time, basis *Basis) bool
