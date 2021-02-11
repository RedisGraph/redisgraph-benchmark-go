package main

import (
	"fmt"
	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/olekukonko/tablewriter"
	"os"
	"sync/atomic"
	"time"
)

type arrayStringParameters []string

func (i *arrayStringParameters) String() string {
	return "my string representation"
}

func (i *arrayStringParameters) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func printFinalSummary(queries []string, queryRates []float64, totalMessages uint64, duration time.Duration) {
	writer := os.Stdout
	messageRate := float64(totalMessages) / float64(duration.Seconds())

	fmt.Printf("\n")
	fmt.Printf("################# RUNTIME STATS #################\n")
	fmt.Printf("Total Duration %.3f Seconds\n", duration.Seconds())
	fmt.Printf("Total Commands issued %d\n", totalCommands)
	fmt.Printf("Total Errors %d ( %3.3f %%)\n", totalErrors, float64(totalErrors/totalCommands*100.0))
	fmt.Printf("Throughput summary: %.0f requests per second\n", messageRate)
	renderGraphResultSetTable(queries, writer, "## Overall RedisGraph resultset stats table\n")
	renderGraphInternalExecutionTimeTable(queries, writer, "## Overall RedisGraph Internal Execution Time summary table\n", serverSide_PerQuery_GraphInternalTime_OverallLatencies, serverSide_AllQueries_GraphInternalTime_OverallLatencies)
	renderTable(queries, writer, "## Overall Client Latency summary table\n", true, true, errorsPerQuery, duration, clientSide_PerQuery_OverallLatencies, clientSide_AllQueries_OverallLatencies)
}

func renderTable(queries []string, writer *os.File, tableTitle string, includeCalls bool, includeErrors bool, errorSlice []uint64, duration time.Duration, detailedHistogram []*hdrhistogram.Histogram, overallHistogram *hdrhistogram.Histogram) {
	fmt.Fprintf(writer, tableTitle)
	data := make([][]string, len(queries)+1)
	for i := 0; i < len(queries); i++ {
		insertTableLine(queries[i], data, i, includeCalls, includeErrors, errorSlice, duration, detailedHistogram[i])
	}
	insertTableLine("Total", data, len(queries), includeCalls, includeErrors, errorSlice, duration, overallHistogram)
	table := tablewriter.NewWriter(writer)
	initialHeader := []string{"Query"}
	if includeCalls {
		initialHeader = append(initialHeader, "Ops/sec")
		initialHeader = append(initialHeader, "Total Calls")
	}
	if includeErrors {
		initialHeader = append(initialHeader, "Total Errors")
	}
	initialHeader = append(initialHeader, "Avg. latency(ms)", "p50 latency(ms)", "p95 latency(ms)", "p99 latency(ms)")
	table.SetHeader(initialHeader)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}
func renderGraphInternalExecutionTimeTable(queries []string, writer *os.File, tableTitle string, detailedHistogram []*hdrhistogram.Histogram, overallHistogram *hdrhistogram.Histogram) {
	fmt.Fprintf(writer, tableTitle)
	initialHeader := []string{"Query", " Internal Avg. latency(ms)", "Internal p50 latency(ms)", "Internal p95 latency(ms)", "Internal p99 latency(ms)"}
	data := make([][]string, len(queries)+1)
	i := 0
	for i = 0; i < len(queries); i++ {
		data[i] = make([]string, 5)
		data[i][0] = queries[i]
		data[i][1] = fmt.Sprintf("%.3f", float64(detailedHistogram[i].Mean()/1000.0))
		data[i][2] = fmt.Sprintf("%.3f", float64(detailedHistogram[i].ValueAtQuantile(50.0))/1000.0)
		data[i][3] = fmt.Sprintf("%.3f", float64(detailedHistogram[i].ValueAtQuantile(95.0))/1000.0)
		data[i][4] = fmt.Sprintf("%.3f", float64(detailedHistogram[i].ValueAtQuantile(99.0))/1000.0)
	}
	data[i] = make([]string, 5)
	data[i][0] = "Total"
	data[i][1] = fmt.Sprintf("%.3f", float64(overallHistogram.Mean()/1000.0))
	data[i][2] = fmt.Sprintf("%.3f", float64(overallHistogram.ValueAtQuantile(50.0))/1000.0)
	data[i][3] = fmt.Sprintf("%.3f", float64(overallHistogram.ValueAtQuantile(95.0))/1000.0)
	data[i][4] = fmt.Sprintf("%.3f", float64(overallHistogram.ValueAtQuantile(99.0))/1000.0)
	table := tablewriter.NewWriter(writer)
	table.SetHeader(initialHeader)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}

func insertTableLine(queryName string, data [][]string, i int, includeCalls, includeErrors bool, errorsSlice []uint64, duration time.Duration, histogram *hdrhistogram.Histogram) {
	data[i] = make([]string, 5)
	latencyPadding := 0
	data[i][0] = queryName
	if includeCalls {
		totalCmds := histogram.TotalCount()
		cmdRate := float64(totalCmds) / float64(duration.Seconds())
		data[i][1] = fmt.Sprintf("%.f", cmdRate)
		data[i][2] = fmt.Sprintf("%d", histogram.TotalCount())
		data[i] = append(data[i], "", "")
		latencyPadding += 2

	}
	if includeErrors {
		var errorV uint64
		// total errors
		if i == (len(data) - 1) {
			errorV = totalErrors
		} else {
			errorV = errorsSlice[i]
		}
		data[i][1+latencyPadding] = fmt.Sprintf("%d", errorV)
		data[i] = append(data[i], "")
		latencyPadding++
	}
	data[i][1+latencyPadding] = fmt.Sprintf("%.3f", float64(histogram.Mean()/1000.0))
	data[i][2+latencyPadding] = fmt.Sprintf("%.3f", float64(histogram.ValueAtQuantile(50.0))/1000.0)
	data[i][3+latencyPadding] = fmt.Sprintf("%.3f", float64(histogram.ValueAtQuantile(95.0))/1000.0)
	data[i][4+latencyPadding] = fmt.Sprintf("%.3f", float64(histogram.ValueAtQuantile(99.0))/1000.0)
}

func renderGraphResultSetTable(queries []string, writer *os.File, tableTitle string) {
	fmt.Fprintf(writer, tableTitle)
	initialHeader := []string{"Query", "Nodes created", "Nodes deleted", "Labels added", "Properties set", " Relationships created", " Relationships deleted"}
	data := make([][]string, len(queries)+1)
	i := 0
	for i = 0; i < len(queries); i++ {
		data[i] = make([]string, 7)
		data[i][0] = queries[i]
		data[i][1] = fmt.Sprintf("%d", totalNodesCreatedPerQuery[i])
		data[i][2] = fmt.Sprintf("%d", totalNodesDeletedPerQuery[i])
		data[i][3] = fmt.Sprintf("%d", totalLabelsAddedPerQuery[i])
		data[i][4] = fmt.Sprintf("%d", totalPropertiesSetPerQuery[i])
		data[i][5] = fmt.Sprintf("%d", totalRelationshipsCreatedPerQuery[i])
		data[i][6] = fmt.Sprintf("%d", totalRelationshipsDeletedPerQuery[i])
	}
	data[i] = make([]string, 7)
	data[i][0] = "Total"
	data[i][1] = fmt.Sprintf("%d", totalNodesCreated)
	data[i][2] = fmt.Sprintf("%d", totalNodesDeleted)
	data[i][3] = fmt.Sprintf("%d", totalLabelsAdded)
	data[i][4] = fmt.Sprintf("%d", totalPropertiesSet)
	data[i][5] = fmt.Sprintf("%d", totalRelationshipsCreated)
	data[i][6] = fmt.Sprintf("%d", totalRelationshipsDeleted)
	table := tablewriter.NewWriter(writer)
	table.SetHeader(initialHeader)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}

func updateCLI(startTime time.Time, tick *time.Ticker, c chan os.Signal, message_limit uint64, loop bool) bool {

	start := startTime
	prevTime := startTime
	prevMessageCount := uint64(0)
	var currentCmds uint64
	var currentErrs uint64
	messageRateTs := []float64{}
	fmt.Printf("%26s %7s %25s %25s %7s %25s %25s %26s\n", "Test time", " ", "Total Commands", "Total Errors", "", "Command Rate", "Client p50 with RTT(ms)", "Graph Internal Time p50 (ms)")
	for {
		select {
		case <-tick.C:
			{
				now := time.Now()
				took := now.Sub(prevTime)
				currentCmds = atomic.LoadUint64(&totalCommands)
				currentErrs = atomic.LoadUint64(&totalErrors)
				messageRate := calculateRateMetrics(int64(currentCmds), int64(prevMessageCount), took)
				completionPercentStr := "[----%]"
				if !loop {
					completionPercent := float64(currentCmds) / float64(message_limit) * 100.0
					completionPercentStr = fmt.Sprintf("[%3.1f%%]", completionPercent)
				}
				errorPercent := float64(currentErrs) / float64(currentCmds) * 100.0

				instantHistogramsResetMutex.Lock()
				p50 := float64(clientSide_AllQueries_OverallLatencies.ValueAtQuantile(50.0)) / 1000.0
				p50RunTimeGraph := float64(serverSide_AllQueries_GraphInternalTime_OverallLatencies.ValueAtQuantile(50.0)) / 1000.0
				instantP50 := float64(clientSide_AllQueries_InstantLatencies.ValueAtQuantile(50.0)) / 1000.0
				instantP50RunTimeGraph := float64(serverSide_AllQueries_GraphInternalTime_InstantLatencies.ValueAtQuantile(50.0)) / 1000.0
				instantHistogramsResetMutex.Unlock()
				if currentCmds != 0 {
					messageRateTs = append(messageRateTs, messageRate)
				}
				prevMessageCount = currentCmds
				prevTime = now

				fmt.Printf("%25.0fs %s %25d %25d [%3.1f%%] %25.2f %19.3f (%3.3f) %20.3f (%3.3f)\t", time.Since(start).Seconds(), completionPercentStr, currentCmds, currentErrs, errorPercent, messageRate, instantP50, p50, instantP50RunTimeGraph, p50RunTimeGraph)
				fmt.Printf("\r")
				if message_limit > 0 && currentCmds >= message_limit && !loop {
					return true
				}
				// The locks we acquire here do not affect the clients
				resetInstantHistograms()
				break
			}

		case <-c:
			fmt.Println("\nReceived Ctrl-c - shutting down cli updater go-routine")
			return false
		}
	}
}
