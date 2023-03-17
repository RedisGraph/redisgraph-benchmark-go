package main

import (
	"log"
	"math"
	"math/rand"
	"strconv"
)

func sample(cdf []float32) int {
	r := rand.Float32()
	bucket := 0
	for (bucket < len(cdf)) && (r > cdf[bucket]) {
		bucket++
	}
	if bucket >= len(cdf) {
		bucket = bucket - 1
	}
	return bucket
}

func prepareCommandsDistribution(queries arrayStringParameters, cmds []string, cmdRates []float64) (int, []float32) {
	var totalDifferentCommands = len(cmds)
	var totalRateSum = 0.0
	var err error
	for i, rawCmdString := range queries {
		cmds[i] = rawCmdString
		if i >= len(benchmarkQueryRates) {
			cmdRates[i] = 1

		} else {
			cmdRates[i], err = strconv.ParseFloat(benchmarkQueryRates[i], 64)
			if err != nil {
				log.Fatalf("Error while converting query-rate param %s: %v", benchmarkQueryRates[i], err)
			}
		}
		totalRateSum += cmdRates[i]
	}
	// probability density function
	if math.Abs(1.0-totalRateSum) > 0.01 {
		log.Fatalf("Total ratio should be 1.0 ( currently is %f )", totalRateSum)
	}
	// probability density function
	if len(benchmarkQueryRates) > 0 && (len(benchmarkQueryRates) != (len(benchmarkQueries) + len(benchmarkQueriesRO))) {
		log.Fatalf("When specifiying -query-rate parameter, you need to have the same number of -query/-query-ro and -query-rate parameters. Number of time -query ( %d ) != Number of times -query-params ( %d )", len(benchmarkQueries), (len(benchmarkQueryRates) + len(benchmarkQueriesRO)))
	}
	pdf := make([]float32, len(queries))
	cdf := make([]float32, len(queries))
	for i := 0; i < len(cmdRates); i++ {
		pdf[i] = float32(cmdRates[i])
		cdf[i] = 0
	}
	// get cdf
	cdf[0] = pdf[0]
	for i := 1; i < len(cmdRates); i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}
	return totalDifferentCommands, cdf
}
