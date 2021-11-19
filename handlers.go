package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const msgAuthFailed = "Authentication Failed"

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
		log.Println("Redirect back After Auth")
		if b1, ac := hasAuthCode(r); b1 {
			// get token from AuthN provider /token
			b2, tk := getIdToken(ac)
			if b2 {
				//log.Printf("Token = %s\n", tk)
				// set as session, http only cookie
				ck := http.Cookie{
					Name:     config().CookieName,
					Value:    tk,
					HttpOnly: false,
				}
				http.SetCookie(w, &ck)
				// forward to backend/upstream -ideal
				//rh.revproxy.ServeHTTP(w, r)

				// XOR ??

				// roundtrip back to self clear browser url!!
				self := fmt.Sprintf("http://%s", r.Host)
				http.Redirect(w, r,
					self,
					http.StatusFound)
			} else {
				// return 401
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(msgAuthFailed))
			}

		} else {
			// return 401
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(msgAuthFailed))
		}

	} else {
		// request from client
		if isAlreadyAuthN(r) {
			//fwd/proxy to backend
			rh.revproxy.ServeHTTP(w, r)
		} else {
			// redirect to AuthN provider
			// send 'self' as return
			rdURL := fmt.Sprintf("%s?return=%s", config().authURL, r.Host)
			http.Redirect(w, r, rdURL, http.StatusFound)
		}
	}
	// debug
	//log.Printf("Req = %v\n", r.Header)
}

func getIdToken(code string) (bool, string) {
	// POSt request to AuthN API
	dt := struct {
		Code string
	}{
		code,
	}
	bd, err := json.Marshal(&dt)
	if err != nil {
		return false, ""
	}
	rsp, err := http.Post(config().tokenURL, "application/json",
		bytes.NewBuffer(bd))
	if err != nil {
		return false, ""
	}
	if rsp.StatusCode == http.StatusOK {
		defer rsp.Body.Close()
		// read response body
		rb, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return false, ""
		}
		sb := string(rb)
		//log.Printf("Resp = %v\n", sb)
		// Assume itis proper token
		// Further validation ignored
		return true, sb
		/*
			curl -X POST http://localhost:8585/api/token \
			 -H 'Content-Type: application/json'  \
			 -d '{"Code": "QWxhbgoyMDIxLTEwLTAzIDE5OjA1OjEwLjc4NDQ5MA=="}'

			>> eyJTdWJqZWN0IjoiQWxhbiIsIlRpbWVTdGFtcCI6IjIwMjEtMTAtMDMgMTk6MDU6MzQuNjY2NDE2In0=
		*/
	}

	return false, ""
}

func hasAuthCode(r *http.Request) (bool, string) {
	// check auth code as query param
	q := r.URL.Query()
	ac, x := q["code"]
	//log.Printf("access code ==> %v\n", ac)
	if x && len(ac) == 1 {
		// try decode auth code
		_, err := b64.URLEncoding.DecodeString(ac[0])
		// cleanup queryparam to remove authcode
		// in further communication
		q.Del("code")
		r.URL.RawQuery = q.Encode()
		//log.Printf("URL After ==> %v\n", r.URL)
		return err == nil, ac[0]
	}
	// if reached here it failed
	return false, ""
}

func isReqFromAuth(r *http.Request) bool {
	// check req origin in header
	o := r.Header.Get("Origin")
	log.Printf("Origin = %s ; AuthUrl = %s\n", o, config().authURL)
	if o != "" {
		if config().AuthPort == "80" || config().AuthPort == "443" {
			// for default ports 'Origin' header will not have port
			// Origin = http://authsimple.com
			// AuthUrl = http://authsimple.com:80
			return o == config().AuthHost
		}
		// else check host & port
		return o == config().authURL
	}
	// if reached here re is not from auth provider
	return false
}

func isAlreadyAuthN(r *http.Request) bool {
	// check if request has idtoken in cookie
	_, err := r.Cookie(config().CookieName)
	return err == nil
}
