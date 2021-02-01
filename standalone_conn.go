package main

import (
	rg "github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
	"log"
)

func getStandaloneConn(graphName, network, addr string, password string) (graph rg.Graph, conn redis.Conn) {
	var err error
	if password != "" {
		conn, err = redis.Dial(network, addr, redis.DialPassword(password))
	} else {
		conn, err = redis.Dial(network, addr)
	}
	if err != nil {
		log.Fatalf("Error preparing for benchmark, while creating new connection. error = %v", err)
	}
	return rg.GraphNew(graphName, conn), conn
}
