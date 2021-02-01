package main

import (
	"flag"
	"fmt"
	"github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/time/rate"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

func main() {
	host := flag.String("h", "127.0.0.1", "Server hostname.")
	port := flag.Int("p", 6379, "Server port.")
	rps := flag.Int64("rps", 0, "Max rps. If 0 no limit is applied and the DB is stressed up to maximum.")
	password := flag.String("a", "", "Password for Redis Auth.")
	clients := flag.Uint64("c", 50, "number of clients.")
	numberRequests := flag.Uint64("n", 1000000, "Total number of requests")
	debug := flag.Int("debug", 0, "Client debug level.")
	randomSeed := flag.Int64("random-seed", 12345, "Random seed to use.")
	graphKey := flag.String("graph-key", "graph", "graph key.")
	flag.Var(&benchmarkQueries, "query", "Specify a RedisGraph query to send in quotes. Each command that you specify is run with its ratio. For example: -query=\"CREATE (n)\" -query-ratio=2")
	flag.Var(&benchmarkQueryRates, "query-ratio", "The query ratio vs other queries used in the same benchmark. Each command that you specify is run with its ratio. For example: -query=\"CREATE (n)\" -query-ratio=10 -query=\"MATCH (n) RETURN n\" -query-ratio=1")
	jsonOutputFile := flag.String("json-out-file", "benchmark-results.json", "Name of json output file to output benchmark results. If not set, will not print to json.")
	//loop := flag.Bool("l", false, "Loop. Run the tests forever.")
	// disabling this for now while we refactor the benchmark client (please use a very large total command number in the meantime )
	// in the meantime added this two fake vars
	var loopV = false
	var loop *bool = &loopV
	flag.Parse()
	if len(benchmarkQueries) < 1 {
		log.Fatalf("You need to specify at least a query with the -query parameter. For example: -query=\"CREATE (n)\"")
	}
	fmt.Printf("Debug level: %d.\n", *debug)
	fmt.Printf("Using random seed: %d.\n", *randomSeed)
	rand.Seed(*randomSeed)
	testResult := NewTestResult("", uint(*clients), *numberRequests, uint64(*rps), "")
	testResult.SetUsedRandomSeed(*randomSeed)

	var requestRate = Inf
	var requestBurst = 1
	useRateLimiter := false
	if *rps != 0 {
		requestRate = rate.Limit(*rps)
		requestBurst = int(*clients)
		useRateLimiter = true
	}

	var rateLimiter = rate.NewLimiter(requestRate, requestBurst)
	samplesPerClient := *numberRequests / *clients
	client_update_tick := 1

	connectionStr := fmt.Sprintf("%s:%d", *host, *port)
	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}
	if !*loop {
		fmt.Printf("Total clients: %d. Commands per client: %d Total commands: %d\n", *clients, samplesPerClient, *numberRequests)
	} else {
		fmt.Printf("Running in loop until you hit Ctrl+C\n")
	}
	queries := make([]string, len(benchmarkQueries))
	cmdRates := make([]int, len(benchmarkQueries))
	totalDifferentCommands, cdf := prepareCommandsDistribution(queries, cmdRates)

	createRequiredGlobalStructs(totalDifferentCommands)

	rgs := make([]redisgraph.Graph, *clients)
	conns := make([]redis.Conn, *clients)

	// a WaitGroup for the goroutines to tell us they've stopped
	dataPointProcessingWg := sync.WaitGroup{}
	graphDatapointsChann := make(chan GraphQueryDatapoint, *numberRequests)

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	c1 := make(chan os.Signal, 1)
	signal.Notify(c1, os.Interrupt)

	tick := time.NewTicker(time.Duration(client_update_tick) * time.Second)

	dataPointProcessingWg.Add(1)
	go processGraphDatapointsChannel(graphDatapointsChann, c1, *numberRequests, &dataPointProcessingWg, &instantHistogramsResetMutex)

	startTime := time.Now()
	for client_id := 0; uint64(client_id) < *clients; client_id++ {
		wg.Add(1)
		rgs[client_id], conns[client_id] = getStandaloneConn(*graphKey, "tcp", connectionStr, *password)
		go ingestionRoutine(&rgs[client_id], true, queries, cdf, samplesPerClient, *loop, *debug, &wg, useRateLimiter, rateLimiter, graphDatapointsChann)
	}

	// enter the update loop
	updateCLI(startTime, tick, c, *numberRequests, *loop)

	endTime := time.Now()
	duration := time.Since(startTime)

	// benchmarked ended, close the connections
	for _, standaloneConn := range conns {
		standaloneConn.Close()
	}

	//wait for all stats to be processed
	dataPointProcessingWg.Wait()

	testResult.FillDurationInfo(startTime, endTime, duration)
	testResult.BenchmarkFullyRun = totalCommands == *numberRequests
	testResult.IssuedCommands = totalCommands
	testResult.OverallGraphInternalQuantiles = GetOverallQuantiles(queries, serverSide_PerQuery_GraphInternalTime_OverallLatencies, serverSide_AllQueries_GraphInternalTime_OverallLatencies)
	testResult.OverallClientQuantiles = GetOverallQuantiles(queries, clientSide_PerQuery_OverallLatencies, clientSide_AllQueries_OverallLatencies)
	testResult.OverallQueryRates = GetOverallRatesMap(duration, queries, clientSide_PerQuery_OverallLatencies, clientSide_AllQueries_OverallLatencies)
	testResult.Totals = GetTotalsMap(queries, clientSide_PerQuery_OverallLatencies, clientSide_AllQueries_OverallLatencies, errorsPerQuery, totalNodesCreatedPerQuery, totalNodesDeletedPerQuery, totalLabelsAddedPerQuery, totalPropertiesSetPerQuery, totalRelationshipsCreatedPerQuery, totalRelationshipsDeletedPerQuery)

	// final merge of pending stats
	printFinalSummary(queries, cmdRates, totalCommands, duration)

	if strings.Compare(*jsonOutputFile, "") != 0 {
		saveJsonResult(testResult, jsonOutputFile)
	}
}
