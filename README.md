# redisgraph-benchmark-go

[![license](https://img.shields.io/github/license/RedisGraph/redisgraph-benchmark-go.svg)](https://github.com/RedisGraph/redisgraph-benchmark-go)
[![GitHub issues](https://img.shields.io/github/release/RedisGraph/redisgraph-benchmark-go.svg)](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest)
[![Discord](https://img.shields.io/discord/697882427875393627?style=flat-square)](https://discord.gg/gWBRT6P)

## Overview

This repo contains code to quick benchmark RedisGraph, using the official [redisgraph-go](https://github.com/RedisGraph/redisgraph-go) client.  

## Installation

### Download Standalone binaries ( no Golang needed )

If you don't have go on your machine and just want to use the produced binaries you can download the following prebuilt bins:

https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest

| OS | Arch | Link |
| :---         |     :---:      |          ---: |
| Linux   | amd64  (64-bit X86)     | [redisgraph-benchmark-go-linux-amd64](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest/download/redisgraph-benchmark-go-linux-amd64.tar.gz)    |
| Linux   | arm64 (64-bit ARM)     | [redisgraph-benchmark-go-linux-arm64](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest/download/redisgraph-benchmark-go-linux-arm64.tar.gz)    |
| Darwin   | amd64  (64-bit X86)     | [redisgraph-benchmark-go-darwin-amd64](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest/download/redisgraph-benchmark-go-darwin-amd64.tar.gz)    |
| Darwin   | arm64 (64-bit ARM)     | [redisgraph-benchmark-go-darwin-arm64](https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest/download/redisgraph-benchmark-go-darwin-arm64.tar.gz)    |

Here's how bash script to download and try it:

```bash
wget -c https://github.com/RedisGraph/redisgraph-benchmark-go/releases/latest/download/redisgraph-benchmark-go-$(uname -mrs | awk '{ print tolower($1) }')-$(dpkg --print-architecture).tar.gz -O - | tar -xz

# give it a try
./redisgraph-benchmark-go --help
```


### Installation in a Golang env

To install the benchmark utility with a Go Env do as follow:

`go get` and then `go install`:
```bash
# Fetch this repo
go get github.com/RedisGraph/redisgraph-benchmark-go
cd $GOPATH/src/github.com/RedisGraph/redisgraph-benchmark-go
make
```

## Usage of redisgraph-benchmark-go

```
$ $ ./redisgraph-benchmark-go --help
Usage of ./redisgraph-benchmark-go:
  -a string
        Password for Redis Auth.
  -c uint
        number of clients. (default 50)
  -continue-on-error
        Continue benchmark in case of error replies.
  -debug int
        Client debug level.
  -enable-exporter-rps
        Push results to redistimeseries exporter in real-time. Time granularity is set via the -reporting-period parameter.
  -exporter-rts-auth string
        RedisTimeSeries Password for Redis Auth.
  -exporter-rts-host string
        RedisTimeSeries hostname. (default "127.0.0.1")
  -exporter-rts-port int
        RedisTimeSeries port. (default 6379)
  -exporter-run-name string
        Run name. (default "perf-run")
  -graph-key string
        graph key. (default "graph")
  -h string
        Server hostname. (default "127.0.0.1")
  -json-out-file string
        Name of json output file to output benchmark results. If not set, will not print to json. (default "benchmark-results.json")
  -n uint
        Total number of requests (default 1000000)
  -p int
        Server port. (default 6379)
  -query value
        Specify a RedisGraph query to send in quotes. Each command that you specify is run with its ratio. For example: -query="CREATE (n)" -query-ratio=1
  -query-ratio value
        The query ratio vs other queries used in the same benchmark. Each command that you specify is run with its ratio. For example: -query="CREATE (n)" -query-ratio=0.5 -query="MATCH (n) RETURN n" -query-ratio=0.5
  -query-ro value
        Specify a RedisGraph read-only query to send in quotes. You can run multiple commands (both read/write) on the same benchmark. Each command that you specify is run with its ratio. For example: -query="CREATE (n)" -query-ratio=0.5 -query-ro="MATCH (n) RETURN n" -query-ratio=0.5
  -random-int-max int
        __rand_int__ upper value limit. __rand_int__ distribution is uniform Random (default 1000000)
  -random-int-min int
        __rand_int__ lower value limit. __rand_int__ distribution is uniform Random (default 1)
  -random-seed int
        Random seed to use. (default 12345)
  -reporting-period duration
        Period to report stats. (default 10s)
  -rps int
        Max rps. If 0 no limit is applied and the DB is stressed up to maximum.
  -v    Output version and exit
```

## Sample output - 100K write commands

```
$ redisgraph-benchmark-go -n 100000 -graph-key graph -query "CREATE (u:User)" 
2021/07/12 11:44:13 redisgraph-benchmark-go (git_sha1:)
2021/07/12 11:44:13 RTS export disabled.
2021/07/12 11:44:13 Debug level: 0.
2021/07/12 11:44:13 Using random seed: 12345.
2021/07/12 11:44:13 Total clients: 50. Commands per client: 2000 Total commands: 100000
2021/07/12 11:44:13 Trying to extract RedisGraph version info
2021/07/12 11:44:13 Detected RedisGraph version 999999

                 Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal Time p50 (ms)
                       10s [100.0%]                    100000                         0 [0.0%]                   9997.46               2.698 (2.698)                2.589 (2.589)       
################# RUNTIME STATS #################
Total Duration 10.004 Seconds
Total Commands issued 100000
Total Errors 0 ( 0.000 %)
Throughput summary: 9996 requests per second
## Overall RedisGraph resultset stats table
|      QUERY      | NODES CREATED | NODES DELETED | LABELS ADDED | PROPERTIES SET | RELATIONSHIPS CREATED  | RELATIONSHIPS DELETED  |
|-----------------|---------------|---------------|--------------|----------------|------------------------|------------------------|
| CREATE (u:User) |        100000 |             0 |            0 |              0 |                      0 |                      0 |
| Total           |        100000 |             0 |            0 |              0 |                      0 |                      0 |
## Overall RedisGraph Internal Execution Time summary table
|      QUERY      | INTERNAL AVG  LATENCY(MS)  | INTERNAL P50 LATENCY(MS) | INTERNAL P95 LATENCY(MS) | INTERNAL P99 LATENCY(MS) |
|-----------------|----------------------------|--------------------------|--------------------------|--------------------------|
| CREATE (u:User) |                      2.599 |                    2.589 |                    2.912 |                    3.648 |
| Total           |                      2.599 |                    2.589 |                    2.912 |                    3.648 |
## Overall Client Latency summary table
|      QUERY      | OPS/SEC | TOTAL CALLS | TOTAL ERRORS | AVG  LATENCY(MS) | P50 LATENCY(MS) | P95 LATENCY(MS) | P99 LATENCY(MS) |
|-----------------|---------|-------------|--------------|------------------|-----------------|-----------------|-----------------|
| CREATE (u:User) |    9996 |      100000 |            0 |            2.745 |           2.698 |           3.048 |           4.007 |
| Total           |    9996 |      100000 |            0 |            2.745 |           2.698 |           3.048 |           4.007 |
2021/07/12 11:44:23 Saving JSON results file to benchmark-results.json
```


## Sample output - running mixed read and writes benchmark

```
$ redisgraph-benchmark-go -n 100000 -graph-key graph -query "CREATE (u:User)" -query-ratio 0.5 -query-ro "MATCH (n) return COUNT(n)" -query-ratio 0.5
2021/07/12 11:45:38 redisgraph-benchmark-go (git_sha1:)
2021/07/12 11:45:38 RTS export disabled.
2021/07/12 11:45:38 Debug level: 0.
2021/07/12 11:45:38 Using random seed: 12345.
2021/07/12 11:45:38 Total clients: 50. Commands per client: 2000 Total commands: 100000
2021/07/12 11:45:38 Trying to extract RedisGraph version info
2021/07/12 11:45:38 Detected RedisGraph version 999999

                 Test time                    Total Commands              Total Errors                      Command Rate   Client p50 with RTT(ms) Graph Internal Time p50 (ms)
                       10s [100.0%]                    100000                         0 [0.0%]                   9996.09               1.179 (1.179)                0.155 (0.155)       
################# RUNTIME STATS #################
Total Duration 10.004 Seconds
Total Commands issued 100000
Total Errors 0 ( 0.000 %)
Throughput summary: 9996 requests per second
## Overall RedisGraph resultset stats table
|           QUERY           | NODES CREATED | NODES DELETED | LABELS ADDED | PROPERTIES SET | RELATIONSHIPS CREATED  | RELATIONSHIPS DELETED  |
|---------------------------|---------------|---------------|--------------|----------------|------------------------|------------------------|
| CREATE (u:User)           |         49921 |             0 |            0 |              0 |                      0 |                      0 |
| MATCH (n) return COUNT(n) |             0 |             0 |            0 |              0 |                      0 |                      0 |
| Total                     |         49921 |             0 |            0 |              0 |                      0 |                      0 |
## Overall RedisGraph Internal Execution Time summary table
|           QUERY           | INTERNAL AVG  LATENCY(MS)  | INTERNAL P50 LATENCY(MS) | INTERNAL P95 LATENCY(MS) | INTERNAL P99 LATENCY(MS) |
|---------------------------|----------------------------|--------------------------|--------------------------|--------------------------|
| CREATE (u:User)           |                      3.825 |                    3.913 |                    4.639 |                    5.249 |
| MATCH (n) return COUNT(n) |                      0.050 |                    0.048 |                    0.067 |                    0.100 |
| Total                     |                      1.935 |                    0.155 |                    4.442 |                    4.929 |
## Overall Client Latency summary table
|           QUERY           | OPS/SEC | TOTAL CALLS | TOTAL ERRORS | AVG  LATENCY(MS) | P50 LATENCY(MS) | P95 LATENCY(MS) | P99 LATENCY(MS) |
|---------------------------|---------|-------------|--------------|------------------|-----------------|-----------------|-----------------|
| CREATE (u:User)           |    4990 |       49921 |            0 |            4.041 |           4.061 |           4.843 |           5.930 |
| MATCH (n) return COUNT(n) |    5006 |       50079 |            0 |            0.236 |           0.178 |           0.442 |           1.201 |
| Total                     |    9996 |      100000 |            0 |            2.135 |           1.179 |           4.611 |           5.287 |
2021/07/12 11:45:48 Saving JSON results file to benchmark-results.json
```
