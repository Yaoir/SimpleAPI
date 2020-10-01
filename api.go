package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	)

// Endpoint: .../api0/previous

func PreviousHandleFunc(w http.ResponseWriter, r *http.Request) {
	var file *os.File		// handle for the data file
	var previous, current int64	// numbers in the sequence
	var result int64		// the number requested
	var status int			// HTTP Status Code

	file, previous, current = readDatafile()
	result, status = handle_previous(&previous,&current)
	writeDatafile(file,previous,current)

	// Return the result.

	if Debug {
		fmt.Printf("%d\n",result)
		fmt.Printf("status: %d\n",status)
		fmt.Printf("%s\n",ToJson(result,status))
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(ToJson(result,status))
}

// Endpoint: .../api0/current

func CurrentHandleFunc(w http.ResponseWriter, r *http.Request) {
	var file *os.File		// handle for the data file
	var previous, current int64	// numbers in the sequence
	var result int64		// the number requested
	var status int			// HTTP Status Code

	file, previous, current = readDatafile()
	result, status = handle_current(&previous,&current)
	writeDatafile(file,previous,current)

	// Return the result.

	if Debug {
		fmt.Printf("%d\n",result)
		fmt.Printf("status: %d\n",status)
		fmt.Printf("%s\n",ToJson(result,status))
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(ToJson(result,status))
}

// Endpoint: .../api0/next

func NextHandleFunc(w http.ResponseWriter, r *http.Request) {
	var file *os.File		// handle for the data file
	var previous, current int64	// numbers in the sequence
	var result int64		// the number requested
	var status int			// HTTP Status Code

	file, previous, current = readDatafile()
	result, status = handle_next(&previous,&current)
	writeDatafile(file,previous,current)

	// Return the result.

	if Debug {
		fmt.Printf("%d\n",result)
		fmt.Printf("status: %d\n",status)
		fmt.Printf("%s\n",ToJson(result,status))
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(ToJson(result,status))
}

func show_memory_usage(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"Total system memory used: %9d bytes. Allocated on heap: %9d bytes.\n", m.Sys, m.HeapAlloc)
}

// All other requests. Return Status Code 400 (Bad Request)

func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w,"<p><h1>HTTP Status Code 400</h1><p><h2>Bad Request from client.</h2></p>")
}

func start_api_server() {
	http.HandleFunc("/",index)	// show "Bad Request" message for any other than the following:
	http.HandleFunc("/test/stats",show_memory_usage)	// report server's memory usage
	// The API endpoints.
	http.HandleFunc("/api0/previous",PreviousHandleFunc)
	http.HandleFunc("/api0/current",CurrentHandleFunc)
	http.HandleFunc("/api0/next",NextHandleFunc)

	p := ":" + strconv.Itoa(Port)
	log.Printf("Starting fibonacci server on port %s\tPID: %d\n", p, os.Getpid())

	// This is a starting place for customizing the server configuation.
	// The timeout values are taken from the documentation of the net/http package.
	s := &http.Server{
		Addr:         p,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}

// eof
