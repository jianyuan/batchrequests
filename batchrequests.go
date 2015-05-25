package batchrequests

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/kr/pretty"
)

type BatchRequestHandler struct {
	serveMux *http.ServeMux
	handler  http.Handler
}

type batchRequestHandler struct {
	handler http.Handler
}

type BatchRequest struct {
	Method string
	URL    string
	Body   string
}

type BatchResponse struct {
	Code    int
	Headers http.Header
	Body    string
}

func New(batchPattern string, handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	brh := &BatchRequestHandler{
		serveMux: http.NewServeMux(),
		handler:  handler,
	}

	brh.serveMux.Handle(batchPattern, &batchRequestHandler{handler})

	return brh
}

func (brh *BatchRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("serving", r)

	var handler http.Handler

	handler, pattern := brh.serveMux.Handler(r)

	if pattern == "" {
		handler = brh.handler
	}

	handler.ServeHTTP(w, r)
}

func (brh *batchRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqs []BatchRequest
	var resps []BatchResponse
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&reqs); err == io.EOF {
		fmt.Fprint(w, "Nothing to decode\n")
	} else if err != nil {
		fmt.Fprintf(w, "Error: %#v\n", err)
	}
	pretty.Fprintf(w, "Got %# v\n", reqs)

	for _, req := range reqs {
		pretty.Fprintf(w, "Trying %# v\n", req)
		newW := httptest.NewRecorder()
		newR, _ := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Body))
		brh.handler.ServeHTTP(newW, newR)
		pretty.Fprintf(w, "Done req %# v, body: %# v\n", newW, newW.Body.String())
		resps = append(resps, BatchResponse{
			Code:    newW.Code,
			Headers: newW.Header(),
			Body:    newW.Body.String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	pretty.Fprintf(w, "Responses %# v\n", resps)
	log.Println(json.NewEncoder(w).Encode(resps))
}
