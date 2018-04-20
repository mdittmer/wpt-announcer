package api

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

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

type Revision struct {
	Hash       string    `json:"hash"`
	CommitTime time.Time `json:"commit_time"`
}

// LatestRequest is models a request for the latest announced revisions.
//
// @jsonschema(
// 	title="Latest revisions request",
//	description="The HTTP get parameters for a request for the latest announced revisions."
// )
//
//go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api LatestRequest
type LatestRequest struct{}

// LatestResponse is models a response for the latest announced revisions.
//
// @jsonschema(
// 	title="Latest revisions response",
//	description="The JSON format for a response containing the latest announced revisions."
// )
//
//go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api LatestResponse
type LatestResponse struct {
	Revisions map[string]Revision `json:"revisions"'`
	Epochs    []Epoch             `json:"epochs"`
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

	rs := make(map[string]Revision)

	for i := range es {
		if len(revs[epochs[i]]) == 0 {
			continue
		}
		rev := revs[epochs[i]][0]
		rs[es[i].ID] = Revision{
			Hash:       fmt.Sprintf("%020x", rev.GetHash()),
			CommitTime: rev.GetCommitTime(),
		}
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

// EpochsRequest is models a request for the epochs supported by the service.
//
// @jsonschema(
// 	title="List-of-epochs request",
//	description="The HTTP GET parameters for a list of the epochs supported by the service."
// )
//go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api EpochsRequest
type EpochsRequest struct{}

// EpochsResponse is models a response for the epochs supported by the service.
//
// @jsonschema(
// 	title="List-of-epochs response",
//	description="The JSON format for a response containing the epochs supported by the service."
// )
////go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api EpochsResponse
type EpochsResponse []Epoch

// RevisionsRequest is models a request for the announced revisions.
//
// @jsonschema(
// 	title="Revisions request",
//	description="The HTTP get parameters for a request for specific announced revisions. Use `epochs` to filter by epochs (default all). Use `num_revisions` to specify number of revisions per epoch (default 1). Use `now` to specify an upper bound on commit time. Use `start` to specify a lower bound on commit time."
// )
//
//go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api RevisionsRequest

type RevisionsRequest struct {
	Epochs       []epoch.Epoch `json:"epochs,omitempty"`
	NumRevisions int           `json:"num_revisions,omitempty"`
	Now          time.Time     `json:"now,omitempty"`
	Start        time.Time     `json:"start,omitempty"`
}

// RevisionsResponse is models a response for the announced revisions.
//
// @jsonschema(
// 	title="Revisions response",
//	description="The JSON format for a response containing announced revisions."
// )
//
//go:generate jsonschemagen github.com/mdittmer/wpt-announcer/api RevisionsResponse
type RevisionsResponse struct {
	Revisions map[string][]Revision `json:"revisions"`
	Epochs    []Epoch               `json:"epochs"`
	Error     string                `json:"error,omitempty"`
}

func RevisionsFromEpochs(revs map[epoch.Epoch][]agit.Revision, apiErr error) RevisionsResponse {
	epochs := make([]epoch.Epoch, 0, len(revs))
	for e := range revs {
		epochs = append(epochs, e)
	}
	sort.Sort(epoch.ByMaxDuration(epochs))
	es := make([]Epoch, 0, len(epochs))
	for _, e := range epochs {
		es = append(es, FromEpoch(e))
	}

	rs := make(map[string][]Revision)

	for i := range es {
		if len(revs[epochs[i]]) == 0 {
			continue
		}
		revs := revs[epochs[i]]
		apiRevs := make([]Revision, 0, len(revs))
		for _, rev := range revs {
			apiRevs = append(apiRevs, Revision{
				Hash:       fmt.Sprintf("%020x", rev.GetHash()),
				CommitTime: rev.GetCommitTime(),
			})
		}
		rs[es[i].ID] = apiRevs
	}

	var response RevisionsResponse
	if apiErr != nil {
		response = RevisionsResponse{
			rs,
			es,
			apiErr.Error(),
		}
	} else {
		response = RevisionsResponse{
			rs,
			es,
			"",
		}
	}

	return response
}
