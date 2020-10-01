package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	)

// The name of the data file.
// It is a 16-byte binary file that holds two 64-bit integers.
// The command line option '-d <filename>' overrides this default.

var datafile = "/tmp/fibonacci_api_data"

// Debug Mode. When true, prints extra information. For development/debugging.

var Debug = false

// TCP/IP port number the server listens on.

var Port = 8080

// The two exceptional conditions that the API can encounter

// Received "previous" request when at the start of the sequence.
const ApiErrorNoPrev = -1
// Received "next" request, and the next number overflows the int64 type.
const ApiErrorOverflow = -2

// Operations associated with API endpoints.

const (
	get_previous = iota
	get_current
	get_next
	)

// The previous and current numbers in the sequence.
// These are written to and restored from the data file.

var nums struct {
	Previous int64
	Current  int64
}

// For the JSON response that is returned to the client.
// "value" is the number in the sequence.
// "status" is the HTTP Status Code.

type Response struct {
	Value	string `json:"value"`
	Status	string `json:"status"`
}

// Convert value (an int64) and status (an int) to JSON.

func ToJson(value int64, status int) []byte {
	var r Response
	r.Value = strconv.FormatInt(value,10)
	r.Status = strconv.FormatInt(int64(status),10)
	jsonstr, err := json.Marshal(r)
	if err != nil {
// TODO: What to do here?
// maybe a custom json error message (500 Internal Server Error)
	}
	return jsonstr
}

// Close the data file.

func closeDatafile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Printf("Could not close data file %s.\n",datafile)
	}
}

// mutex for locking access to the data file
var mutex = &sync.Mutex{}

// Read previous and current from the data file.
// Return both of those and a file handle.
// The file handle is returned so it can be written to and closed in writeDatafile()

func readDatafile() (*os.File, int64, int64) {
	var file *os.File
	var previous, current int64
	var err error
	var restart bool			// true if the API resets the sequence to the start (0)

	// Lock the data file.
	// It stays locked until just after it is written and closed.
	mutex.Lock()

	// Open the data file. If it doesn't exist, create it.
	file, err = os.OpenFile(datafile, os.O_RDWR|os.O_CREATE, 0600)	// read and write only by owner
	if err != nil {
		log.Printf("Fatal - Cannot open data file %s for writing.\n",datafile)
		os.Exit(2)
	}

	err = binary.Read(file,binary.BigEndian,&nums)
	// If the file is not found or contains (0,0), initialize it with (-1,0).
	if err != nil || (nums.Previous == 0 && nums.Current == 0) {
		restart = true
		log.Printf("Initializing data file.")
	}

	previous = nums.Previous
	current = nums.Current
	if Debug { fmt.Printf("Read from data file:  %d %d\n",previous,current) }

	if restart {
		previous = -1
		current = 0
		if Debug { fmt.Printf("Resetting to: %d %d\n",previous,current) }
	}

	return file, previous, current
}

// Write previous and current to the data file.

func writeDatafile(file *os.File, previous int64, current int64) {
	nums.Previous = previous
	nums.Current = current
	if Debug { fmt.Printf("Writing to data file: %d %d\n",nums.Previous,nums.Current) }

	// Write the file starting at the first byte.
	file.Seek(0,0)
	err := binary.Write(file,binary.BigEndian,&nums)
	if err != nil {
		log.Printf("Error writing data file %s\n",datafile)
		// The server's state is not inconsistent so force a restart.
		os.Exit(2)
	}
	file.Close()
// TODO: if err != nil {} ?
	// Unlock the data file so other threads can access it.
	mutex.Unlock()
}

// The following three functions handle the previous, current, and next operatoins.
// Each returns result (the number requested) and status (the HTTP Status Code)

func handle_previous(previous *int64, current *int64) (int64, int) {
	var result int64
	var next int64
	var status int

	if *previous == -1 {
		result = ApiErrorNoPrev
		status = http.StatusConflict
		if Debug { fmt.Printf("API Error: cannot step backwards.\n") }
		// Write this to the data file:
		*previous = -1
		*current = 0
	} else if *previous == 0 && *current == 1 {
		result = 0
		status = http.StatusOK
		if Debug { fmt.Printf("Previous: %d\n",0) }
		*previous = -1
		*current = 0
	} else {
		result = *previous
		status = http.StatusOK
		if Debug { fmt.Printf("Previous: %d\n",*previous) }
		// Step backwards to find the new previous to save in the data file.
		next = *current - *previous
		*current = *previous
		*previous = next
	}

	return result, status
}

func handle_current(previous *int64, current *int64) (int64, int) {
	var result int64
	var status int

	result = *current
	status = http.StatusOK

	return result,status
}

func handle_next(previous *int64, current *int64) (int64, int) {
	var result int64
	var next int64
	var status int

	// Handle initial condition.
	if *previous == -1 {
		// Write this to the data file:
		*previous = 0
		*current = 1
		result = 1
		status = http.StatusOK
		if Debug { fmt.Printf("Next: %d\n",1) }
	} else {
		// Find the next Fibonacci number
		next = *previous + *current
		// There was an overflow if next is negative.
		if next < 0 {
			result = ApiErrorOverflow
			status = http.StatusInsufficientStorage
			if Debug { fmt.Printf("API Error: cannot step forward.\n") }
		} else {
			result = next
			status = http.StatusOK
			if Debug { fmt.Printf("Next: %d\n",next) }
			// step forward for writing the data file.
			*previous, *current = *current, next
		}
	}

	return result, status
}

/* The CLI handler */

func cli_handle_request(op int) {
	var file *os.File		// handle for the data file
	var previous, current int64	// numbers in the sequence
	var result int64		// the number requested
	var status int			// HTTP Status Code

	file, previous, current = readDatafile()

	switch op {
		case get_previous:	// step backwards
			result, status = handle_previous(&previous,&current)
		case get_current:	// return current number in sequence
			result, status = handle_current(&previous,&current)
		case get_next:		// step forwards
			result, status = handle_next(&previous,&current)
	}

	writeDatafile(file,previous,current)

	// Print the result.

	fmt.Printf("result: %d\n",result)
	fmt.Printf("status: %d\n",status)
	fmt.Printf("JSON: %s\n",ToJson(result,status))
}

func main() {
	var op int		// API operation code (get_previous, get_current, or get_next)
	var cli bool = false	// true if running as interactive shell command, not HTTP server
	var cputest = false	// If running 1 CPU test

	// Handle command line arguments.

	flag.StringVar(&datafile,"f",datafile,"Name of data file.")	// -f <name>
	flag.IntVar(&Port,"p",Port,"TCP/IP port number.")		// -p <port_num>
	flag.BoolVar(&Debug,"d",Debug,"Print debug messages.")		// -d
	flag.BoolVar(&cputest,"t",false,"Limit to 1 CPU.")		// -t
	flag.Parse()

	// If there are any non-option arguments, run "interactively" as a command.

	if len(flag.Args()) != 0 { cli = true }

	// -t flag: Limit the Go runtime to using only 1 CPU.
	if cputest { runtime.GOMAXPROCS(1) }

	// Check the arguments.

	if Port > 65535 || Port < 0 {
		if cli {
			fmt.Fprintf(os.Stderr,"bad port number %d\n",Port)
			os.Exit(1)
		} else {
			log.Fatalf("bad port number %d\n",Port)
			os.Exit(1)
		}
	}

	if datafile == "" {
		if cli {
			fmt.Fprintf(os.Stderr,"Null name for data file\n")
			os.Exit(1)
		} else {
			log.Fatalf("Null name for data file\n")
			os.Exit(1)
		}
	}

	// If there are no non-option arguments, run as server.

	if ! cli {
		start_api_server()
		os.Exit(0)
	}

	// Otherwise, run as a command.
	// One argument is allowed, which must be 'p', 'c', or 'n'

	if(len(flag.Args()) != 1) {
		log.Printf("bad arguments\n")
		os.Exit(1)
	}

	switch flag.Arg(0) {
		case "p": op = get_previous
		case "c": op = get_current
		case "n": op = get_next
		default:
			fmt.Printf("bad argument: %s\n",flag.Arg(0))
			os.Exit(1)
	}

	cli_handle_request(op)
}

