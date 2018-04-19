package api

import (
	"errors"
	"reflect"
	"sort"

	"github.com/mdittmer/wpt-announcer/epoch"
	agit "github.com/mdittmer/wpt-announcer/git"
	strcase "github.com/stoewer/go-strcase"
)

var errMissingRevision = errors.New("Missing required revision")

func GetErMissingRevision() error {
	return errMissingRevision
}

type Epoch struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	MinDuration float32 `json:"min_duration_sec"`
	MaxDuration float32 `json:"max_duration_sec"`
}

func FromEpoch(e epoch.Epoch) Epoch {
	t := reflect.TypeOf(e)
	v := reflect.ValueOf(e)
	for t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}
	id := strcase.SnakeCase(t.Name())
	d := e.GetData()
	minDuration := float32(d.MinDuration.Seconds())
	maxDuration := float32(d.MaxDuration.Seconds())
	return Epoch{
		id,
		d.Label,
		d.Description,
		minDuration,
		maxDuration,
	}
}

type LatestResponse struct {
	Revisions map[string]agit.Revision
	Epochs    []Epoch
}

func LatestFromEpochs(revs map[epoch.Epoch][]agit.Revision) (LatestResponse, error) {
	epochs := make([]epoch.Epoch, 0, len(revs))
	for e := range revs {
		epochs = append(epochs, e)
	}
	sort.Sort(epoch.ByMaxDuration(epochs))
	es := make([]Epoch, 0, len(epochs))
	for _, e := range epochs {
		es = append(es, FromEpoch(e))
	}

	rs := make(map[string]agit.Revision)

	for i := range es {
		if len(revs[epochs[i]]) == 0 {
			continue
		}
		rs[es[i].ID] = revs[epochs[i]][0]
	}

	latest := LatestResponse{
		rs,
		es,
	}

	if len(rs) < len(epochs) {
		return latest, errMissingRevision
	}

	return latest, nil
}

type EpochsResponse []Epoch
