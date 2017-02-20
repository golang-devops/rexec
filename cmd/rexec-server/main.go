package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	_ "github.com/golang-devops/rexec/comms"
	"github.com/golang-devops/rexec/logging"
)

var (
	address = flag.String("address", "127.0.0.1:50505", "The address to listen on")
)

func main() {
	logger := logging.Logger()

	flag.Parse()
	if *address == "" {
		logger.Fatal("address flag is missing")
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info(fmt.Sprintf("Server listening on %s", *address))
	if err := http.Serve(listener, nil); err != nil {
		logger.Fatal(err.Error())
	}
}
