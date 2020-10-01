## Status

This is an initial pre-release version, and is not intended to be used for a production server.

This is the first release, for internal team review and discussion only.

I am not considering this done yet, and am submitting it for the purpose of asking questions. I have another important job interview coming up that I need to prepare for, so I will ask questions now, wait for a response, and continue to finish this later, probably after I do that other interview.

Reviewer(s): Please read the Assumptions and Questions sections, and let me know if I need to do anything differently. I will then make appropriate modifications and resubmit.

#### Assumptions

1. Since the spec referred to HTTP protocol and not HTTPS, I did not overly concern myself with security.

2. For the implementation, I assume you want me to just write a server program that implements the API, and I do not need to go any farther, such as Dockerizing it and adding Kubernetes and a cloud server.

If I made any of the above assumptions incorrectly, please let me know.

#### Questions

1. Will a 64-bit integer be sufficient for holding numbers in the sequence,
   or should I use arbitrary precision?
   There is a large performance tradeoff with supporting arbitrary-precision
   numbers, especially with very large numbers in the sequence. (I do not know for
   sure that this can be done while still meeting the 1000 request per second spec.)
   The benefit of arbitrary precision is that the sequence can continue up to an
   arbitrary limit, rather than being limited to integers that fit in 64 bits.
   Aribitrary precision in Go can be added with the **math/big** package.
   https://golang.org/pkg/math/big/

2. Is there a preference for how to handle overflow (the next in the sequence
   being higher than a pre-defined limit) and underflow (the previous number
   being before zero, the start of the sequence)? The method I am using is
   described in the subsection titled "Error Conditions" in this README file.

3. The specification did not say how the API should behave in the case of multiple clients using it.

   Should there be one sequence per client, so each client is fully independent of the others?
   (That is, each client's view of the sequence is private, and its sequence is not modified
   by other clients.)

   Or should all clients access/modify the same sequence? In this case, any client can
   step forward or backward in the sequence, and the change will affect other clients.

   To keep the initial implementation simple, I implemented the latter for now.

   The former can be implemented using a cookie and a larger data file, with a record for each active client.
   I have already thought this through somewhat, and it would not be really difficult to add.
   But it would lower performance, with both lower throughput and greater memory footprint.

4. There are other things I can add or modify if you want. Let me know if you find anything is missing.

## Introduction

This repository is a response to the programming exercise specified at

https://gist.github.com/DuoVR/7febcd39aa0e1be2b18f44d163685e4b

## Building the Executable

The API server is written in Go, using nothing outside of the Go standard library. To compile it, you need to have Go installed. Go to [https://golang.org/dl/](https://golang.org/dl/) to download and install Go.

To compile:

```
$ go build -o fibserver fib.go api.go
```
or if you have GNU **make** installed:
```
$ make
```

The makefile contains many other targets that are helpful for development.

## Running the Server

$ fibserver

#### Command Options and Arguments

```
-p <port_num>
	Specify TCP/IP port number to listen on.

-f <filename>
	Specify name of data file. Include the full path specification.
	Example: /tmp/fibbonaci_data_file

-d
	Turn on Debug mode. Prints extra information.
	For debugging and/or development only.

-t
	Test mode, which forces the Go runtime to use only one CPU thread.
	This simulates performance on a machine with only one CPU.

p, c, n
	If any non-option arguments are provided, it runs in interactive shell mode,
	performing one of the previous, current, or next operations. In this case,
	the -p option, if provided, is ignored.
```

### Usage Examples

```
$ fibserver p
```

prints the previous number in the sequence.

```
$ fibserver c
```

prints the current number in the sequence.

```
$ fibserver n
```

prints the next number in the sequence.

```
$ fibserver -p 5000 -f /var/run/fibserver.data 2>/var/log/fibserver.log
```

runs as a server on port 5000, using /var/run/fibserver.data as the data file, and with log messages written to /var/log/fibserver.log. (Note: /run typically uses the tmpfs volatile filesystem on Linux. Do not use this example if you need the API to survive operating system reboots. See the section Surviving Server Restarts and OS Reboots later in this document.)

## API Documentation

The API returns numbers in the Fibonacci sequence.
https://en.wikipedia.org/wiki/Fibonacci_number

#### API Endpoints

The API has three endpoints:

```
/api0/next - Return the next number in the sequence.
/api0/current - Return the current number in the sequence.
/api0/previous - Return the previous number in the sequence.
```

A request to any of the endpoints returns a JSON object like this:

```
{"value":"13","status":"200"}
```

**value** is the requested number from the sequence, and **status** is the HTTP Status Code in the response from the server.

#### Error Conditions

If already the start of the sequence and the **previous** endpoint is requested, the response from the server is:

```
{"value":"-1","status":"409"}
```

If a request for the **next** endpoint is made that results in the overflow of the 64-bit integer used to hold the numbers in the sequence, the server responds with:

```
{"value":"-2","status":"507"}
```

In both of the above cases, the **current** endpoint will continue to function, returning the start of the sequence or the largest number possible, and the **next** and **previous** endpoints will function as previously described.

## Implementation Notes

Go was used because performance (throughput) was one of the requirements, and Go is one of the languages used at Pex.
The only practical way to gain any more performance (throughput) would be to write it in C, which would be take longer. Also, C is not type safe nor memory safe, and the performance gain would be only about 10-20%, at the most.

I did not implement a full REST (RESTful) API because it was not a requirement (or even mentioned), and because this API is so simple.

### Surviving Server Restarts and OS Reboots

The spec included, "The API should also be able to recover and restart if it unexpectedly crashes."

The method used to ensure survivability of the API across server restarts and operating system reboots (whether they are done intentionally or as the result of a crash) is to store the current state of the API, as two consecutive values of the sequence, in a data file. The data file is a tiny 16-byte binary file, with 16 bytes holding two (8-byte) 64-bit signed integers. The integers are stored most significant byte first. The first number is the **previous** value in the sequence, and the second number is the **current** value.

This method was used because it has good data integrity (like a database), while being far more efficient.

If surviving reboots is desired, make sure to put the data file in a location that is stored on a *persistent storage device*, such as a hard drive or SSD, and not in a directory that exists only in memory while the system is running (that is, *volatile memory* such as **/proc** or any directory in a **tmpfs** filesystem on Linux), or a temporary directory that is emptied when the system boots (such as **/tmp**). If the data file is held in volatile memory, the API can survive server restarts, but not operating system reboots.

I probably do not need to add this, but if completely fault-tolerant operation were required, an even more robust implementation would be needed. As it is, if the operating system crashes while a request is being processed, the resulting state of the data file may not match what the API returns to clients.

#### File Locking

A Go mutex (https://golang.org/pkg/sync/#Mutex) is used to lock the data file for concurrent access by the threads. Since OS-level file locking is not used, care must be taken to ensure that only one **fibserver** process is running on the host system.

## Performance

The spec asked for an API that could handle 1000 requests per second, and run on a "small machine" with one CPU and 512 MB of memory.

### Throughput

Since minimum throughput was listed first in the requirements, the API server was designed with that as a top priority.
It was specified that the server needs to run on a 1 CPU machine. 

The server called **runtime.MAXPROCS(1)** to limit the Go runtime to using only 1 CPU thread.

#### Client Side

Throughput was measured using **wrk** http server testing tool.

https://github.com/wg/wrk

```
The following **wrk** options were used for the measurement:
        -t3: using 3 threads for clients
        -c1000: maximum of 1000 connections
        -d100s: testing for 100 seconds
```

Here is the result:

```
$ wrk -t3 -c1000 -d100s http://localhost:8080/api0/current
Running 2m test @ http://localhost:8080/api0/current
  3 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    83.68ms   21.05ms 173.23ms   83.14%
    Req/Sec     4.00k   419.76     5.30k    62.57%
  1193250 requests in 1.67m, 171.83MB read
Requests/sec:  11922.88
Transfer/sec:      1.72MB
```

This was on an Intel Core i3 machine running at 3 GHz, with the data file stored on a solid state drive (SSD) mounted on the root of the filesystem. (Roughly similar to servers commonly used in datacenters.) If running on a Raspberry Pi 2 (a low-power ARM-based machine), performance would be about 20% of this, which would still be well over 1000 requests per second.

The above test showed around 12 thousand requests per second when the data file was stored on a SSD. Tests were also done with the data file residing in a filesystem on a hard drive and in volatile memory. Here are the results:

```
Storage Type	Requests Per Second
------------	-------------------
Hard Drive	16222
SSD		11922
Volatile	15738
```

Notice that the hard drive produced the highest throughput, even higher than when using a data file in a volatile (RAM-based) filesystem. This shows that factors other than the underlying storage technology are very significant. In actual practice, a test needs to be done on the actual server (such as a cloud server or dedicated server in a datacenter) used as the production server.

### Memory Usage

An endpoint was added for obtaining memory usage statistics from the running server:

```
/test/stats
```

This endpoint was requested during the above througput test to measure memory usage on a loaded server.
The server responded with the following:

```
Total system memory used:  74990592 bytes. Allocated on heap:  24655056 bytes.
```

While running the test on the development machine, total memory used remained constant at exactly 74990592 bytes (about 75 MB), showing that a server with 512 MB of main memory (RAM) is more than sufficient, and there are no memory leaks. The "Total system memory used" figure measures the total amount of memory consumed "by the Go runtime for the heap, stacks, and other internal data structures". See the Go **runtime** package documentation for details.  https://golang.org/pkg/runtime/#MemStats

## Additional Notes

The TODO file in the repository contains a list of what I would do next if I continued to work on this.
