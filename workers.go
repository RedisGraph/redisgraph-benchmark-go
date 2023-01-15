package main

import (
	"fmt"
	"github.com/RedisGraph/redisgraph-go"
	"golang.org/x/time/rate"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func ingestionRoutine(rg *redisgraph.Graph, continueOnError bool, cmdS []string, commandIsRO []bool, commandsCDF []float32, randomIntPadding, randomIntMax int64, number_samples uint64, loop bool, debug_level int, wg *sync.WaitGroup, useLimiter bool, rateLimiter *rate.Limiter, statsChannel chan GraphQueryDatapoint, replacementEnabled bool, replacementArr []map[string]string, commandStartPos uint64) {
	defer wg.Done()
	var replacementTerms map[string]string
	for i := 0; uint64(i) < number_samples || loop; i++ {
		cmdPos := sample(commandsCDF)
		termReplacementPos := commandStartPos + uint64(i)
		if replacementEnabled {
			replacementTerms = replacementArr[termReplacementPos]
		}
		sendCmdLogic(rg, cmdS[cmdPos], commandIsRO[cmdPos], randomIntPadding, randomIntMax, cmdPos, continueOnError, debug_level, useLimiter, rateLimiter, statsChannel, replacementEnabled, replacementTerms)
	}
}

func sendCmdLogic(rg *redisgraph.Graph, query string, readOnly bool, randomIntPadding, randomIntMax int64, cmdPos int, continueOnError bool, debug_level int, useRateLimiter bool, rateLimiter *rate.Limiter, statsChannel chan GraphQueryDatapoint, replacementEnabled bool, replacementTerms map[string]string) {
	if useRateLimiter {
		r := rateLimiter.ReserveN(time.Now(), int(1))
		time.Sleep(r.Delay())
	}
	var err error
	var queryResult *redisgraph.QueryResult

	processedQuery := processQuery(query, randomIntPadding, randomIntMax, replacementEnabled, replacementTerms)
	startT := time.Now()
	if readOnly {
		queryResult, err = rg.ROQuery(processedQuery)
	} else {
		queryResult, err = rg.Query(processedQuery)
	}
	endT := time.Now()

	duration := endT.Sub(startT)
	datapoint := GraphQueryDatapoint{
		CmdPos:                      cmdPos,
		ClientDurationMicros:        duration.Microseconds(),
		GraphInternalDurationMicros: 0,
		Error:                       false,
		Empty:                       true,
		NodesCreated:                0,
		NodesDeleted:                0,
		LabelsAdded:                 0,
		PropertiesSet:               0,
		RelationshipsCreated:        0,
		RelationshipsDeleted:        0,
	}
	if err != nil {
		datapoint.Error = true
		if continueOnError {
			if debug_level > 0 {
				log.Println(fmt.Sprintf("Received an error with the following query(s): %v, error: %v", query, err))
			}
		} else {
			log.Fatalf("Received an error with the following query(s): %v, error: %v", query, err)
		}
	} else {
		datapoint.GraphInternalDurationMicros = int64(queryResult.InternalExecutionTime() * 1000.0)
		if debug_level > 1 {
			fmt.Printf("Issued query: %s\n", query)
			fmt.Printf("Pretty printing result:\n")
			queryResult.PrettyPrint()
			fmt.Printf("\n")
		}
		datapoint.Empty = queryResult.Empty()
		datapoint.NodesCreated = uint64(queryResult.NodesCreated())
		datapoint.NodesDeleted = uint64(queryResult.NodesDeleted())
		datapoint.LabelsAdded = uint64(queryResult.LabelsAdded())
		datapoint.PropertiesSet = uint64(queryResult.PropertiesSet())
		datapoint.RelationshipsCreated = uint64(queryResult.RelationshipsCreated())
		datapoint.RelationshipsDeleted = uint64(queryResult.RelationshipsDeleted())
	}
	statsChannel <- datapoint
}

func processQuery(query string, randomIntPadding int64, randomIntMax int64, replacementEnabled bool, replacementTerms map[string]string) string {
	if replacementEnabled {
		for placeholder, term := range replacementTerms {
			query = strings.Replace(query, placeholder, term, -1)
		}
	}
	for strings.Index(query, randIntPlaceholder) != -1 {
		randIntString := fmt.Sprintf("%d", rand.Int63n(randomIntMax)+randomIntPadding)
		query = strings.Replace(query, randIntPlaceholder, randIntString, 1)
	}
	return query
}
