package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// handler funcs for diff req
type reqHandler struct {
	revproxy *httputil.ReverseProxy
}

func (rh *reqHandler) init() {
	// initilaize reverse proxy with bckend urlto host
	url, _ := url.Parse(config().backendURL)
	rh.revproxy = httputil.NewSingleHostReverseProxy(url)
}

func (rh *reqHandler) finalize() {
	// do nothing for now
}

// handler func for /hello[?name=xxx]
func (rh *reqHandler) hello(w http.ResponseWriter, r *http.Request) {
	q, ok := r.URL.Query()["name"]
	if ok {
		fmt.Fprintf(w, "<h1>Salut, Bonjour  %s!</h1>", string(q[0]))
	} else {
		fmt.Fprint(w, "<h1>Salut, Bonjour!</h1>")
	}
}

// handler func for /datetime => JSON
func (rh *reqHandler) datetime(w http.ResponseWriter, r *http.Request) {
	n := time.Now()
	dt := struct {
		Date string
		Time string
	}{
		n.Format("2006 Jan 02"),
		n.Format("03:04:05 PM"),
	}
	js, err := json.Marshal(dt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (rh *reqHandler) basehandler(w http.ResponseWriter, r *http.Request) {
	if isReqFromAuth(r) {
		// req is a redirect back from auth provider
		if hasValidAuthCode(r) {
			// get token from AuthN provider /token
			// set as session, http only cookie
		} else {
			// AuthN failed !!
			// return 401
		}

	} else {
		if isAlreadyAuthN(r) {
			//fwd/proxy to backend
		} else {
			// redirect to AuthN provider
		}
	}
	// debug
	log.Printf("Req = %v\n", r.Header)
	http.Redirect(w, r, config().authURL, http.StatusFound)
	// forward to backend/upstream
	//rh.revproxy.ServeHTTP(w, r)
}

func hasValidAuthCode(r *http.Request) bool {
	// check req origin in header
	return false
}

func isReqFromAuth(r *http.Request) bool {
	// check req origin in header
	return false
}

func isAlreadyAuthN(r *http.Request) bool {
	// check if request has idtoken in cookie
	return false
}
