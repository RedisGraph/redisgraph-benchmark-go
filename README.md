# redisgraph-benchmark-go

## Overview

This repo contains code to quick benchmark RedisGraph, using the official redisgraph-go client.  

## Installation

The easiest way to get and install the redisgraph-benchmark-go Go program is to use
`go get` and then `go install`:
```bash
# Fetch this repo
go get github.com/filipecosta90/redisgraph-benchmark-go
cd $GOPATH/src/github.com/filipecosta90/redisgraph-benchmark-go
make
```

## Usage of redisgraph-benchmark-go

```
$ ./redisgraph-benchmark-go --help
  Usage of ./redisgraph-benchmark-go:
    -a string
          Password for Redis Auth.
    -c uint
          number of clients. (default 50)
    -debug int
          Client debug level.
    -graph-key string
          graph key. (default "graph")
    -h string
          Server hostname. (default "127.0.0.1")
    -l    Loop. Run the tests forever.
    -n uint
          Total number of requests (default 1000000)
    -p int
          Server port. (default 6379)
    -rps int
          Max rps. If 0 no limit is applied and the DB is stressed up to maximum.
```

## Sample output - 10M commands

```
$ redisgraph-benchmark-go -graph-key graph "CREATE (u:User)"
  Total clients: 50. Commands per client: 20000 Total commands: 1000000
                   Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal p50 with RTT(ms)
                         43s [100.0%]                   1000000                         0 [0.0%]                  14810.59                     2.087                     0.000    
  #################################################
  Total Duration 43.000 Seconds
  Total Errors 0
  Throughput summary: 23256 requests per second
  Overall Client Latency summary (msec):
            p50       p95       p99
          2.087     2.977     3.611
  Overall RedisGraph Internal Execution time Latency summary (msec):
            p50       p95       p99
          0.000     0.000     0.000
```


## Sample output - running in loop mode ( Ctrl+c to stop )

```
$ redisgraph-benchmark-go -l -graph-key graph "CREATE (u:User)"
  Running in loop until you hit Ctrl+C
                   Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal p50 with RTT(ms)
  ^C                        9s [----%]                    263185                         0 [0.0%]                  24167.81                     1.846                     0.000   
  received Ctrl-c - shutting down
  
  #################################################
  Total Duration 9.000 Seconds
  Total Errors 0
  Throughput summary: 29243 requests per second
  Overall Client Latency summary (msec):
            p50       p95       p99
          1.846     2.383     3.117
  Overall RedisGraph Internal Execution time Latency summary (msec):
            p50       p95       p99
          0.000     0.000     0.000
```