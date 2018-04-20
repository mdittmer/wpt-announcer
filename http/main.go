package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
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

func prepareJsonResponse(w http.ResponseWriter, r *http.Request, q url.Values) {
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

var defaultErrorJson = []byte("{\n\t\"error\": \"Unknown error\"\n}")

func strToErrorJson(str string) []byte {
	payload := make(map[string]string)
	payload["error"] = str
	bytes, err := marshal(payload)
	if err != nil {
		return defaultErrorJson
	}
	return bytes
}

type apiData struct {
	basePath     string
	requestType  reflect.Type
	responseType reflect.Type
	handler      func(w http.ResponseWriter, r *http.Request)
}

func (a apiData) schemaHandler(t reflect.Type) func(w http.ResponseWriter, r *http.Request) {
	pkg := strings.Replace(strings.Replace(t.PkgPath(), "/", "-", -1), ".", "_", -1)
	name := t.Name()
	path := fmt.Sprintf("./api/schema/%s-%s.json", pkg, name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", path, err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		prepareJsonResponse(w, r, url.Values{})

		h := w.Header()
		_, ok := h["Link"]
		if !ok {
			h["Link"] = make([]string, 0, 1)
		}

		u := getBaseURLBuffer(r)
		// TODO(markdittmer): This inappropriately exploits the fact that both schema suffixes are the same length.
		implURL := u.String()
		implURL = implURL[:len(implURL)-len(apiRequestSchemaSuffix)]
		h["Link"] = append(h["Link"], fmt.Sprintf("<%s>; rel=\"implementation\"", implURL))

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
		bytes, err := json.Marshal(r.URL.Query())
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJson("Failed validate query parameters"))
			return
		}
		res, err := schema.Validate(gojsonschema.NewBytesLoader(bytes))
		if err != nil {
			w.WriteHeader(500)
			w.Write(strToErrorJson(fmt.Sprintf("Failed validate query parameters: %v", err)))
			return
		}
		if !res.Valid() {
			w.WriteHeader(500)
			w.Write(strToErrorJson(fmt.Sprintf("Failed validate query parameters: %v", res.Errors())))
			return
		}
		h(w, r)
	}
}

func (a apiData) register() {
	http.HandleFunc(a.basePath, a.handler)
	http.HandleFunc(a.basePath+apiRequestSchemaSuffix, a.schemaHandler(a.requestType))
	http.HandleFunc(a.basePath+apiResponseSchemaSuffix, a.schemaHandler(a.responseType))
}

func epochsHandler(w http.ResponseWriter, r *http.Request) {
	prepareJsonResponse(w, r, url.Values{})
	bytes, err := marshal(apiEpochs)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJson("Failed to marshal epochs JSON"))
		return
	}
	w.Write(bytes)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	prepareJsonResponse(w, r, url.Values{})
	if a == nil {
		w.WriteHeader(503)
		w.Write(strToErrorJson("Announcer not yet initialized"))
		return
	}

	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write(strToErrorJson("No epochs"))
		return
	}

	now := time.Now()
	revs, err := a.GetRevisions(latestGetRevisions, announcer.Limits{
		Now:   now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJson(err.Error()))
		return
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJson(err.Error()))
		return
	}

	bytes, err := marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(strToErrorJson("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}

func init() {
	for _, e := range epochs {
		apiEpochs = append(apiEpochs, api.FromEpoch(e))
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

	apis := []apiData{
		apiData{
			"/api/revisions/epochs",
			reflect.TypeOf(api.EpochsRequest{}),
			reflect.TypeOf(api.EpochsResponse{}),
			epochsHandler,
		},
		apiData{
			"/api/revisions/latest",
			reflect.TypeOf(api.LatestRequest{}),
			reflect.TypeOf(api.LatestResponse{}),
			latestHandler,
		},
	}

	for _, a := range apis {
		a.register()
	}

	log.Printf("INFO: Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
