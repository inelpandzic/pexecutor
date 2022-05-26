package main

import (
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/inelpandzic/pexecutor/executor"
	"github.com/inelpandzic/pexecutor/server"
)

const (
	defaultPort      = 8080
	defaulPoolSize   = 5
	defaultQueueSize = 1000
)

var poolSize int
var queueSize int
var port int

func main() {
	flag.IntVar(&poolSize, "pool-size", defaulPoolSize, "Worker pool size")
	flag.IntVar(&queueSize, "queue-size", defaultQueueSize, "Executor task queue size")
	flag.IntVar(&port, "port", defaultPort, "Port number")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	ex := executor.New(poolSize, queueSize, logger)
	go ex.Run()
	defer ex.Close()

	server := server.New(port, ex, logger)
	log.Fatal(server.Serve())
}
