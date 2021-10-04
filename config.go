package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

const configFile = "./config.json"
const portEnvKey = "PORT"

type configData struct {
	Port        string
	BackendHost string
	BackendPort string
	AuthHost    string
	AuthPort    string
	TokenPath   string
	backendURL  string
	authURL     string
	tokenURL    string
	CookieName  string
}

var once sync.Once
var cf *configData

func config() *configData {
	if cf == nil {
		once.Do(
			func() {
				cf = &configData{}
				cf.load()
				cf.backendURL = formURL(cf.BackendHost, cf.BackendPort)
				cf.authURL = formURL(cf.AuthHost, cf.AuthPort)
				cf.tokenURL = fmt.Sprintf("%s/%s", cf.authURL, cf.TokenPath)
			})
	}
	return cf
}

func formURL(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func (cf *configData) load() {
	bytes, err := ioutil.ReadFile(configFile)
	checkerr(err)
	err = json.Unmarshal(bytes, cf)
	// <TO DO> chang unmarshall to map[string]string
	// iterate and check each key is loaded to not empty
	// assign to 'cf' fields
	//log.Printf("cf = %v\n", cf)
	checkerr(err)
	port := os.Getenv(portEnvKey)
	if port != "" {
		cf.Port = port
	}
}
