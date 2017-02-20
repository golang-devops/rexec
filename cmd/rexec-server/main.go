package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"

	_ "github.com/golang-devops/rexec/comms"
)

var (
	address = flag.String("address", "127.0.0.1:50505", "The address to listen on")
)

func main() {
	flag.Parse()
	if *address == "" {
		log.Fatal("address flag is missing")
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("Server listening on %s", *address))
	if err := http.Serve(listener, nil); err != nil {
		log.Fatal(err)
	}
}
