package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/golang-devops/rexec/comms"
)

var (
	serverAddress = flag.String("server-address", "127.0.0.1:50505", "The address of the server")
)

func main() {
	flag.Parse()
	if *serverAddress == "" {
		log.Fatal("server-address flag is missing")
	}

	client, err := comms.NewConnectedClient(*serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	args := &comms.ExecutorExecuteArgs{Exe: "ping", Args: []string{"google.com"}}

	/*executeReply := &comms.ExecutorExecuteReply{}
	if err := client.Execute(args, executeReply); err != nil {
		log.Fatal(err)
	} else if executeReply.Error != nil {
		log.Fatal(executeReply.Error)
	}*/

	startReply := &comms.ExecutorStartReply{}
	if err := client.Start(args, startReply); err != nil {
		log.Fatal(err)
	} else if startReply.Error != nil {
		log.Fatal(startReply.Error)
	}

	// fmt.Println(fmt.Sprintf("OUT: %s", string(executeReply.Out)))
	fmt.Println(fmt.Sprintf("PID: %d", startReply.Pid))
}
