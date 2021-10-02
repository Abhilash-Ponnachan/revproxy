package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// multiplexer to route req to hanlder functions
	muxHandler := http.NewServeMux()

	// struct to encapsulate handler methods
	rh := reqHandler{}
	// initailize handler
	rh.init()
	defer rh.finalize()

	// handle base path
	muxHandler.HandleFunc("/", rh.basehandler)

	// dummy api for testing html response
	muxHandler.HandleFunc("/api/hello", rh.hello)
	// dummy api for testing json reponse
	muxHandler.HandleFunc("/api/datetime", rh.datetime)

	// http server instance
	server := http.Server{
		Addr: fmt.Sprintf(":%s", config().Port),
		// our mux as handler for http requests
		Handler:      muxHandler,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 10,
	}
	log.Printf("Starting http server on PORT: %s\n", config().Port)
	// start server and listen for req to serve
	// this is a 'blocking' call
	err := server.ListenAndServe()
	// check type of termination error
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

}
