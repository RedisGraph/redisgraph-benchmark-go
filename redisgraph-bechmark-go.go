package main

import (
	"flag"
	"fmt"
	hdrhistogram "github.com/HdrHistogram/hdrhistogram-go"
	"github.com/mediocregopher/radix/v3"
	"github.com/redislabs/redisgraph-go"
	"golang.org/x/time/rate"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var totalCommands uint64
var totalEmptyResultsets uint64
var totalErrors uint64

var totalNodesCreated uint64
var totalNodesDeleted uint64
var totalLabelsAdded uint64
var totalPropertiesSet uint64
var totalRelationshipsCreated uint64
var totalRelationshipsDeleted uint64

var latencies *hdrhistogram.Histogram
var graphRunTimeLatencies *hdrhistogram.Histogram

const Inf = rate.Limit(math.MaxFloat64)

func ingestionRoutine(rg redisgraph.Graph, continueOnError bool, cmdS string, number_samples uint64, loop bool, debug_level int, wg *sync.WaitGroup, useLimiter bool, rateLimiter *rate.Limiter) {
	defer wg.Done()
	for i := 0; uint64(i) < number_samples || loop; i++ {
		sendCmdLogic(rg, cmdS, continueOnError, debug_level, useLimiter, rateLimiter)
	}
}

func sendCmdLogic(rg redisgraph.Graph, query string, continueOnError bool, debug_level int, useRateLimiter bool, rateLimiter *rate.Limiter) {
	if useRateLimiter {
		r := rateLimiter.ReserveN(time.Now(), int(1))
		time.Sleep(r.Delay())
	}
	var err error
	var queryResult *redisgraph.QueryResult
	startT := time.Now()
	queryResult, err = rg.Query(query)
	endT := time.Now()
	atomic.AddUint64(&totalCommands, uint64(1))
	duration := endT.Sub(startT)
	if err != nil {
		if continueOnError {
			atomic.AddUint64(&totalErrors, uint64(1))
			if debug_level > 0 {
				log.Println(fmt.Sprintf("Received an error with the following query(s): %v, error: %v", query, err))
			}
		} else {
			log.Fatalf("Received an error with the following query(s): %v, error: %v", query, err)
		}
	} else {
		err = graphRunTimeLatencies.RecordValue(int64(queryResult.RunTime() * 1000))
		if err != nil {
			log.Fatalf("Received an error while recording RedisGraph RunTime latencies: %v", err)
		}
		if debug_level > 1 {
			fmt.Printf("Issued query: %s\n", query)
			fmt.Printf("Pretty printing result:\n")
			queryResult.PrettyPrint()
			fmt.Printf("\n")
		}
		if queryResult.Empty() {
			atomic.AddUint64(&totalEmptyResultsets, uint64(1))
		}
		atomic.AddUint64(&totalNodesCreated, uint64(queryResult.NodesCreated()))
		atomic.AddUint64(&totalNodesDeleted, uint64(queryResult.NodesDeleted()))
		atomic.AddUint64(&totalLabelsAdded, uint64(queryResult.LabelsAdded()))
		atomic.AddUint64(&totalPropertiesSet, uint64(queryResult.PropertiesSet()))
		atomic.AddUint64(&totalRelationshipsCreated, uint64(queryResult.RelationshipsCreated()))
		atomic.AddUint64(&totalRelationshipsDeleted, uint64(queryResult.RelationshipsDeleted()))
	}
	err = latencies.RecordValue(duration.Microseconds())
	if err != nil {
		log.Fatalf("Received an error while recording latencies: %v", err)
	}
}

func main() {
	host := flag.String("h", "127.0.0.1", "Server hostname.")
	port := flag.Int("p", 6379, "Server port.")
	rps := flag.Int64("rps", 0, "Max rps. If 0 no limit is applied and the DB is stressed up to maximum.")
	password := flag.String("a", "", "Password for Redis Auth.")
	clients := flag.Uint64("c", 50, "number of clients.")
	numberRequests := flag.Uint64("n", 1000000, "Total number of requests")
	debug := flag.Int("debug", 0, "Client debug level.")
	loop := flag.Bool("l", false, "Loop. Run the tests forever.")
	graphKey := flag.String("graph-key", "graph", "graph key.")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("You need to specify a query after the flag command arguments.")
	}
	fmt.Printf("Debug level: %d.\n", *debug)

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
	latencies = hdrhistogram.New(1, 90000000, 3)
	graphRunTimeLatencies = hdrhistogram.New(1, 90000000, 3)
	opts := make([]radix.DialOpt, 0)
	if *password != "" {
		opts = append(opts, radix.DialAuthPass(*password))
	}
	connectionStr := fmt.Sprintf("%s:%d", *host, *port)
	stopChan := make(chan struct{})
	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}
	if !*loop {
		fmt.Printf("Total clients: %d. Commands per client: %d Total commands: %d\n", *clients, samplesPerClient, *numberRequests)
	} else {
		fmt.Printf("Running in loop until you hit Ctrl+C\n")
	}
	query := strings.Join(args, " ")

	for channel_id := 1; uint64(channel_id) <= *clients; channel_id++ {
		wg.Add(1)
		cmd := make([]string, len(args))
		copy(cmd, args)
		go ingestionRoutine(getStandaloneConn(*graphKey, "tcp", connectionStr), true, query, samplesPerClient, *loop, *debug, &wg, useRateLimiter, rateLimiter)
	}

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tick := time.NewTicker(time.Duration(client_update_tick) * time.Second)
	closed, _, duration, totalMessages, _ := updateCLI(tick, c, *numberRequests, *loop)
	messageRate := float64(totalMessages) / float64(duration.Seconds())
	p50IngestionMs := float64(latencies.ValueAtQuantile(50.0)) / 1000.0
	p95IngestionMs := float64(latencies.ValueAtQuantile(95.0)) / 1000.0
	p99IngestionMs := float64(latencies.ValueAtQuantile(99.0)) / 1000.0

	graph_p50IngestionMs := float64(graphRunTimeLatencies.ValueAtQuantile(50.0)) / 1000.0
	graph_p95IngestionMs := float64(graphRunTimeLatencies.ValueAtQuantile(95.0)) / 1000.0
	graph_p99IngestionMs := float64(graphRunTimeLatencies.ValueAtQuantile(99.0)) / 1000.0

	fmt.Printf("\n")
	fmt.Printf("################# RUNTIME STATS #################\n")
	fmt.Printf("Total Duration %.3f Seconds\n", duration.Seconds())
	fmt.Printf("Total Commands issued %d\n", totalCommands)
	fmt.Printf("Total Errors %d ( %3.3f %%)\n", totalErrors, float64(totalErrors/totalCommands*100.0))
	fmt.Printf("Throughput summary: %.0f requests per second\n", messageRate)
	fmt.Printf("Overall Client Latency summary (msec):\n")
	fmt.Printf("    %9s %9s %9s\n", "p50", "p95", "p99")
	fmt.Printf("    %9.3f %9.3f %9.3f\n", p50IngestionMs, p95IngestionMs, p99IngestionMs)
	fmt.Printf("################## GRAPH STATS ##################\n")
	fmt.Printf("Total Empty resultsets %d ( %3.3f %%)\n", totalEmptyResultsets, float64(totalEmptyResultsets/totalCommands*100.0))
	fmt.Printf("Total Nodes created %d\n", totalNodesCreated)
	fmt.Printf("Total Nodes deleted %d\n", totalNodesDeleted)
	fmt.Printf("Total Labels added %d\n", totalLabelsAdded)
	fmt.Printf("Total Properties set %d\n", totalPropertiesSet)
	fmt.Printf("Total Relationships created %d\n", totalRelationshipsCreated)
	fmt.Printf("Total Relationships deleted %d\n", totalRelationshipsDeleted)
	fmt.Printf("Overall RedisGraph Internal Execution time Latency summary (msec):\n")
	fmt.Printf("    %9s %9s %9s\n", "p50", "p95", "p99")
	fmt.Printf("    %9.3f %9.3f %9.3f\n", graph_p50IngestionMs, graph_p95IngestionMs, graph_p99IngestionMs)

	if closed {
		return
	}

	// tell the goroutine to stop
	close(stopChan)
	// and wait for them both to reply back
	wg.Wait()
}

func updateCLI(tick *time.Ticker, c chan os.Signal, message_limit uint64, loop bool) (bool, time.Time, time.Duration, uint64, []float64) {

	start := time.Now()
	prevTime := time.Now()
	prevMessageCount := uint64(0)
	messageRateTs := []float64{}
	fmt.Printf("%26s %7s %25s %25s %7s %25s %25s %25s\n", "Test time", " ", "Total Commands", "Total Errors", "", "Command Rate", "Client p50 with RTT(ms)", "Graph Internal p50 with RTT(ms)")
	for {
		select {
		case <-tick.C:
			{
				now := time.Now()
				took := now.Sub(prevTime)
				messageRate := float64(totalCommands-prevMessageCount) / float64(took.Seconds())
				completionPercentStr := "[----%]"
				if !loop {
					completionPercent := float64(totalCommands) / float64(message_limit) * 100.0
					completionPercentStr = fmt.Sprintf("[%3.1f%%]", completionPercent)
				}
				errorPercent := float64(totalErrors) / float64(totalCommands) * 100.0

				p50 := float64(latencies.ValueAtQuantile(50.0)) / 1000.0
				p50RunTimeGraph := float64(graphRunTimeLatencies.ValueAtQuantile(50.0)) / 1000.0

				if prevMessageCount == 0 && totalCommands != 0 {
					start = time.Now()
				}
				if totalCommands != 0 {
					messageRateTs = append(messageRateTs, messageRate)
				}
				prevMessageCount = totalCommands
				prevTime = now

				fmt.Printf("%25.0fs %s %25d %25d [%3.1f%%] %25.2f %25.3f %25.3f\t", time.Since(start).Seconds(), completionPercentStr, totalCommands, totalErrors, errorPercent, messageRate, p50, p50RunTimeGraph)
				fmt.Printf("\r")
				if message_limit > 0 && totalCommands >= uint64(message_limit) && !loop {
					return true, start, time.Since(start), totalCommands, messageRateTs
				}

				break
			}

		case <-c:
			fmt.Println("\nreceived Ctrl-c - shutting down")
			return true, start, time.Since(start), totalCommands, messageRateTs
		}
	}
}
