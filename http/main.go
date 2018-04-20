package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdittmer/wpt-announcer/announcer"
	"github.com/mdittmer/wpt-announcer/api"
	"github.com/mdittmer/wpt-announcer/epoch"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/time/rate"

	"log"

	agit "github.com/mdittmer/wpt-announcer/git"
)

var a announcer.Announcer

var epochs = []epoch.Epoch{
	epoch.Weekly{},
	epoch.Daily{},
	epoch.EightHourly{},
	epoch.FourHourly{},
	epoch.TwoHourly{},
	epoch.Hourly{},
}

var apiEpochs = make([]api.Epoch, 0)

var epochsMap = make(map[string]epoch.Epoch)

var latestGetRevisions = make(map[epoch.Epoch]int)

const (
	apiRequestSchemaSuffix  = "/schema/req"
	apiResponseSchemaSuffix = "/schema/res"
)

func getBaseURLBuffer(r *http.Request) bytes.Buffer {
	ru := r.URL
	var u bytes.Buffer
	if r.TLS == nil {
		u.WriteString("http://")
	} else {
		u.WriteString("https://")
	}
	if ru.User != nil {
		u.WriteString(ru.User.String())
	}
	u.WriteString(r.Host)
	u.WriteString(ru.Path)
	return u
}

func prepareJSONResponse(w http.ResponseWriter, r *http.Request, q url.Values) {
	selfLink := getBaseURLBuffer(r)
	if len(q) > 0 {
		selfLink.WriteString("?")
		selfLink.WriteString(q.Encode())
	}
	h := w.Header()
	h["Content-Type"] = []string{"application/vnd.restful+json"}

	_, ok := h["Link"]
	if !ok {
		h["Link"] = make([]string, 0, 1)
	}
	h["Link"] = append(h["Link"], fmt.Sprintf("<%s>; rel=\"self\"", selfLink.String()))
}

func marshal(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "\t")
}

var defaultErrorJSON = []byte("{\n\t\"error\": \"Unknown error\"\n}")

func strToErrorJSON(str string) []byte {
	payload := make(map[string]string)
	payload["error"] = str
	bytes, err := marshal(payload)
	if err != nil {
		return defaultErrorJSON
	}
	return bytes
}

type apiData struct {
	basePath string
	request  interface{}
	response interface{}
	handler  func(w http.ResponseWriter, r *http.Request)
}

func (a apiData) schemaHandler(t reflect.Type) func(w http.ResponseWriter, r *http.Request) {
	pkg := strings.Replace(strings.Replace(t.PkgPath(), "/", "-", -1), ".", "_", -1)
	name := t.Name()
	path := fmt.Sprintf("../api/schema/%s-%s.json", pkg, name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", path, err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		prepareJSONResponse(w, r, url.Values{})

		h := w.Header()
		_, ok := h["Link"]
		if !ok {
			h["Link"] = make([]string, 0, 1)
		}

		u := getBaseURLBuffer(r)
		// TODO(markdittmer): This inappropriately exploits the fact that both schema suffixes are the same length.
		implURL := u.String()
		implURL = implURL[:len(implURL)-len(apiRequestSchemaSuffix)]
		h["Link"] = append(h["Link"], fmt.Sprintf("<%s>; rel=\"describes\"", implURL))

		log.Print(h)

		w.WriteHeader(200)
		w.Write(bytes)
	}
}

func (a apiData) implHandler(req interface{}, h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(req))
	if err != nil {
		log.Fatalf("Failed to load JSON schema for %v", req)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		prepareJSONResponse(w, r, q)
		hdr := w.Header()
		if _, ok := hdr["Link"]; !ok {
			hdr["Link"] = make([]string, 0, 1)
		}
		u := getBaseURLBuffer(r)
		baseLink := (&u).String()
		hdr["Link"] = append(hdr["Link"], fmt.Sprintf("<%s>; rel=\"describedby\"", baseLink+apiRequestSchemaSuffix))
		hdr["Link"] = append(hdr["Link"], fmt.Sprintf("<%s>; rel=\"describedby\"", baseLink+apiResponseSchemaSuffix))

		bytes, err := json.Marshal(q)
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Failed validate query parameters"))
			return
		}
		res, err := schema.Validate(gojsonschema.NewBytesLoader(bytes))
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJSON(fmt.Sprintf("Failed validate query parameters: %v", err)))
			return
		}
		if !res.Valid() {
			w.WriteHeader(500)
			w.Write(strToErrorJSON(fmt.Sprintf("Failed validate query parameters: %v", res.Errors())))
			return
		}
		h(w, r)
	}
}

func (a apiData) register() {
	http.HandleFunc(a.basePath, a.implHandler(a.request, a.handler))
	http.HandleFunc(a.basePath+apiRequestSchemaSuffix, a.schemaHandler(reflect.TypeOf(a.request)))
	http.HandleFunc(a.basePath+apiResponseSchemaSuffix, a.schemaHandler(reflect.TypeOf(a.response)))
}

func epochsHandler(w http.ResponseWriter, r *http.Request) {
	bytes, err := marshal(apiEpochs)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON("Failed to marshal epochs JSON"))
		return
	}
	w.Write(bytes)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	if a == nil {
		w.WriteHeader(503)
		w.Write(strToErrorJSON("Announcer not yet initialized"))
		return
	}

	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write(strToErrorJSON("No epochs"))
		return
	}

	now := time.Now()
	revs, err := a.GetRevisions(latestGetRevisions, announcer.Limits{
		Now:   now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON(err.Error()))
		return
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON(err.Error()))
		return
	}

	bytes, err := marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}

func revisionsHandler(w http.ResponseWriter, r *http.Request) {
	if a == nil {
		w.WriteHeader(503)
		w.Write(strToErrorJSON("Announcer not yet initialized"))
		return
	}

	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write(strToErrorJSON("No epochs"))
		return
	}

	q := r.URL.Query()

	numRevisions := 1
	if nr, ok := q["num_revisions"]; ok {
		if len(nr) > 1 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Multiple num_revisions values"))
			return
		}
		if len(nr) == 0 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Empty num_revisions value"))
			return
		}
		var err error
		numRevisions, err = strconv.Atoi(nr[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJSON(fmt.Sprintf("Invalid num_revisions value: %s", nr[0])))
			return
		}
	}

	getRevisions := latestGetRevisions
	if eStrs, ok := q["epochs"]; ok {
		getRevisions = make(map[epoch.Epoch]int)
		for _, eStr := range eStrs {
			if e, ok := epochsMap[eStr]; ok {
				getRevisions[e] = numRevisions
			} else {
				w.WriteHeader(500)
				w.Write(strToErrorJSON(fmt.Sprintf("Unknown epoch: %s", eStr)))
				return
			}
		}
	}

	es := make([]epoch.Epoch, 0, len(getRevisions))
	for e := range getRevisions {
		es = append(es, e)
	}
	sort.Sort(epoch.ByMaxDuration(es))

	now := time.Now()
	if tStrs, ok := q["now"]; ok {
		if len(tStrs) > 1 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Multiple now values"))
			return
		}
		if len(tStrs) == 0 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Empty now value"))
			return
		}
		var err error
		now, err = time.Parse("", tStrs[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJSON(fmt.Sprintf("Invalid now value: %s", tStrs[0])))
			return
		}
	}

	start := now.Add(-2 * epochs[0].GetData().MaxDuration)
	if tStrs, ok := q["start"]; ok {
		if len(tStrs) > 1 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Multiple start values"))
			return
		}
		if len(tStrs) == 0 {
			w.WriteHeader(500)
			w.Write(strToErrorJSON("Empty start value"))
			return
		}
		var err error
		now, err = time.Parse("", tStrs[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJSON(fmt.Sprintf("Invalid start value: %s", tStrs[0])))
			return
		}
	}

	revs, err := a.GetRevisions(getRevisions, announcer.Limits{
		Now:   now,
		Start: start,
	})
	if revs == nil && err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON(err.Error()))
		return
	}

	response := api.RevisionsFromEpochs(revs, err)
	bytes, err := marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJSON("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}

func init() {
	for _, e := range epochs {
		apiEpoch := api.FromEpoch(e)
		apiEpochs = append(apiEpochs, apiEpoch)
		epochsMap[apiEpoch.ID] = e
		latestGetRevisions[e] = 1
	}

	go func() {
		log.Print("INFO: Initializing announcer")
		var err error
		a, err = announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
			URL:                       "https://github.com/w3c/web-platform-tests.git",
			RemoteName:                "origin",
			ReferenceName:             "refs/heads/master",
			EpochReferenceIterFactory: announcer.NewBoundedMergedPRIterFactory(),
			Git: agit.GoGit{},
		})
		if err != nil {
			log.Fatalf("Announcer initialization failed: %v", err)
		}
		log.Print("INFO: Announcer initialized")
	}()

	go func() {
		limit := rate.Limit(1.0 / 60.0)
		burst := 1
		limiter := rate.NewLimiter(limit, burst)
		ctx := context.Background()

		for i := 0; true; i++ {
			err := limiter.Wait(ctx)
			if err != nil {
				log.Printf("WARN: Announcer update rate limiter error: %v", err)
			}
			if a == nil {
				log.Print("WARN: Periodic announcer update: Skipping iteration: Announcer not yet initialized")
				continue
			}

			log.Print("INFO: Periodic announcer update: Updating...")
			err = a.Update()
			if err != nil {
				log.Printf("ERRO: Error updating announcer: %v", err)
			}
			log.Print("INFO: Update complete")
		}
	}()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile | log.LUTC)

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(dir)

	apis := []apiData{
		apiData{
			"/api/revisions/epochs",
			api.EpochsRequest{},
			api.EpochsResponse{},
			epochsHandler,
		},
		apiData{
			"/api/revisions/latest",
			api.LatestRequest{},
			api.LatestResponse{},
			latestHandler,
		},
		apiData{
			"/api/revisions/list",
			api.RevisionsRequest{},
			api.RevisionsResponse{},
			revisionsHandler,
		},
	}

	for _, a := range apis {
		a.register()
	}

	log.Printf("INFO: Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
