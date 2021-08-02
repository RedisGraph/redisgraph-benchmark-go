package main

import (
	"github.com/HdrHistogram/hdrhistogram-go"
	"golang.org/x/time/rate"
	"math"
	"sync"
)

var totalCommands uint64
var totalEmptyResultsets uint64
var totalErrors uint64
var errorsPerQuery []uint64

var totalNodesCreated uint64
var totalNodesDeleted uint64
var totalLabelsAdded uint64
var totalPropertiesSet uint64
var totalRelationshipsCreated uint64
var totalRelationshipsDeleted uint64

var totalNodesCreatedPerQuery []uint64
var totalNodesDeletedPerQuery []uint64
var totalLabelsAddedPerQuery []uint64
var totalPropertiesSetPerQuery []uint64
var totalRelationshipsCreatedPerQuery []uint64
var totalRelationshipsDeletedPerQuery []uint64

var randIntPlaceholder string = "__rand_int__"

// no locking is required when using the histograms. data is duplicated on the instant and overall histograms
var clientSide_AllQueries_OverallLatencies *hdrhistogram.Histogram
var serverSide_AllQueries_GraphInternalTime_OverallLatencies *hdrhistogram.Histogram

var clientSide_PerQuery_OverallLatencies []*hdrhistogram.Histogram
var serverSide_PerQuery_GraphInternalTime_OverallLatencies []*hdrhistogram.Histogram

// this mutex does not affect any of the client go-routines ( it's only to sync between main thread and datapoints processer go-routines )
var instantHistogramsResetMutex sync.Mutex
var clientSide_AllQueries_InstantLatencies *hdrhistogram.Histogram
var serverSide_AllQueries_GraphInternalTime_InstantLatencies *hdrhistogram.Histogram

var benchmarkQueries arrayStringParameters
var benchmarkQueriesRO arrayStringParameters
var benchmarkQueryRates arrayStringParameters

const Inf = rate.Limit(math.MaxFloat64)

func createRequiredGlobalStructs(totalDifferentCommands int) {
	errorsPerQuery = make([]uint64, totalDifferentCommands)
	totalNodesCreatedPerQuery = make([]uint64, totalDifferentCommands)
	totalNodesDeletedPerQuery = make([]uint64, totalDifferentCommands)
	totalLabelsAddedPerQuery = make([]uint64, totalDifferentCommands)
	totalPropertiesSetPerQuery = make([]uint64, totalDifferentCommands)
	totalRelationshipsCreatedPerQuery = make([]uint64, totalDifferentCommands)
	totalRelationshipsDeletedPerQuery = make([]uint64, totalDifferentCommands)

	clientSide_AllQueries_OverallLatencies = hdrhistogram.New(1, 90000000000, 4)
	clientSide_AllQueries_InstantLatencies = hdrhistogram.New(1, 90000000000, 4)
	serverSide_AllQueries_GraphInternalTime_OverallLatencies = hdrhistogram.New(1, 90000000000, 4)
	serverSide_AllQueries_GraphInternalTime_InstantLatencies = hdrhistogram.New(1, 90000000000, 4)

	clientSide_PerQuery_OverallLatencies = make([]*hdrhistogram.Histogram, totalDifferentCommands)
	serverSide_PerQuery_GraphInternalTime_OverallLatencies = make([]*hdrhistogram.Histogram, totalDifferentCommands)
	for i := 0; i < totalDifferentCommands; i++ {
		clientSide_PerQuery_OverallLatencies[i] = hdrhistogram.New(1, 90000000000, 4)
		serverSide_PerQuery_GraphInternalTime_OverallLatencies[i] = hdrhistogram.New(1, 90000000000, 4)
	}
}

func resetInstantHistograms() {
	instantHistogramsResetMutex.Lock()
	clientSide_AllQueries_InstantLatencies.Reset()
	serverSide_AllQueries_GraphInternalTime_InstantLatencies.Reset()
	instantHistogramsResetMutex.Unlock()
}
