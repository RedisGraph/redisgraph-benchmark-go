package main

import (
	"crypto/tls"
	"crypto/x509"
	rg "github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"log"
)

func getStandaloneConn(graphName, network, addr string, password string, tlsCaCertFile string) (graph rg.Graph, conn redis.Conn) {

	var err error
	if tlsCaCertFile != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(tlsCaCertFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		clientTLSConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		// InsecureSkipVerify controls whether a client verifies the
		// server's certificate chain and host name.
		// If InsecureSkipVerify is true, TLS accepts any certificate
		// presented by the server and any host name in that certificate.
		// In this mode, TLS is susceptible to man-in-the-middle attacks.
		// This should be used only for testing.
		clientTLSConfig.InsecureSkipVerify = true
		if password != "" {
			conn, err = redis.Dial(network, addr,
				redis.DialPassword(password),
				redis.DialTLSConfig(clientTLSConfig),
				redis.DialUseTLS(true),
				redis.DialTLSSkipVerify(true),
			)
		} else {
			conn, err = redis.Dial(network, addr,
				redis.DialTLSConfig(clientTLSConfig),
				redis.DialUseTLS(true),
				redis.DialTLSSkipVerify(true),
			)
		}
	} else {
		if password != "" {
			conn, err = redis.Dial(network, addr, redis.DialPassword(password))
		} else {
			conn, err = redis.Dial(network, addr)
		}
	}
	if err != nil {
		log.Fatalf("Error preparing for benchmark, while creating new connection. error = %v", err)
	}
	return rg.GraphNew(graphName, conn), conn
}
