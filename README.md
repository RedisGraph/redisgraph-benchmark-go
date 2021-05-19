# redisgraph-benchmark-go

[![license](https://img.shields.io/github/license/RedisGraph/redisgraph-benchmark-go.svg)](https://github.com/RedisGraph/redisgraph-benchmark-go)
[![GitHub issues](https://img.shields.io/github/release/RedisGraph/redisgraph-benchmark-go.svg)](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest)
[![Discord](https://img.shields.io/discord/697882427875393627?style=flat-square)](https://discord.gg/gWBRT6P)

## Overview

This repo contains code to quick benchmark RedisGraph, using the official [redisgraph-go](https://github.com/RedisGraph/redisgraph-go) client.  
The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

## Installation


### Standalone binaries ( no Golang needed )

If you don't have Go on your machine and just want to use the produced binaries you can download 
the prebuilt bins for Windows, Linux, and Darwin from the [Releases page](https://github.com/RedisGraph/redisgraph-benchmark-go/releases).

Here's an example on how to use the above links:
```bash
# Fetch this repo
wget https://github.com/RedisGraph/redisgraph-benchmark-go/releases/download/v0.0.1/redisgraph-benchmark-go_linux_amd64

# change permissions
chmod 755 redisgraph-benchmark-go_linux_amd64

# give it a try 
./redisgraph-benchmark-go_linux_amd64 --help
```


## Installation in a Golang env

The easiest way to get and install the redisgraph-benchmark-go Go program is to use
`go get` and then `go install`:
```bash
# Fetch this repo
go get github.com/RedisGraph/redisgraph-benchmark-go
cd $GOPATH/src/github.com/RedisGraph/redisgraph-benchmark-go
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

## Sample output - 1M commands

```
$ redisgraph-benchmark-go -graph-key graph "CREATE (u:User)"
  Debug level: 0.
  Total clients: 50. Commands per client: 20000 Total commands: 1000000
                   Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal p50 with RTT(ms)
                         53s [100.0%]                   1000000                         0 [0.0%]                   8002.89                     2.097                     0.000    
  ################# RUNTIME STATS #################
  Total Duration 53.001 Seconds
  Total Commands issued 1000000
  Total Errors 0 ( 0.000 %)
  Throughput summary: 18868 requests per second
  Overall Client Latency summary (msec):
            p50       p95       p99
          2.097     5.347     9.063
  ################## GRAPH STATS ##################
  Total Empty resultsets 1000000 ( 100.000 %)
  Total Nodes created 1000000
  Total Nodes deleted 0
  Total Labels added 0
  Total Properties set 0
  Total Relationships created 0
  Total Relationships deleted 0
  Overall RedisGraph Internal Execution time Latency summary (msec):
            p50       p95       p99
          0.000     0.000     0.000
```


## Sample output - running in loop mode ( Ctrl+c to stop )

```
$ redisgraph-benchmark-go -l -graph-key graph "CREATE (:Rider {name:'A'})-[:rides]->(:Team {name:'Z'})"
  Debug level: 0.
  Running in loop until you hit Ctrl+C
                   Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal p50 with RTT(ms)
  ^C                     11s [----%]                    136649                         0 [0.0%]                   7854.48                     3.667                     0.000     
  received Ctrl-c - shutting down
  
  ################# RUNTIME STATS #################
  Total Duration 11.516 Seconds
  Total Commands issued 140704
  Total Errors 0 ( 0.000 %)
  Throughput summary: 12217 requests per second
  Overall Client Latency summary (msec):
            p50       p95       p99
          3.751     6.887     8.623
  ################## GRAPH STATS ##################
  Total Empty resultsets 140705 ( 100.000 %)
  Total Nodes created 281410
  Total Nodes deleted 0
  Total Labels added 0
  Total Properties set 281410
  Total Relationships created 140705
  Total Relationships deleted 0
  Overall RedisGraph Internal Execution time Latency summary (msec):
            p50       p95       p99
          0.000     0.000     0.000
```
