package main

import (
	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	"log"
)

func getStandaloneConn(graphName, network, addr string) rg.Graph {
	conn, err := redis.Dial(network, addr)
	if err != nil {
		log.Fatalf("Error preparing for benchmark, while creating new connection. error = %v", err)
	}
	graph := rg.GraphNew(graphName, conn)
	return graph
}
