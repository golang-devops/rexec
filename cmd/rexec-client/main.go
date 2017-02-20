package main

import (
	"fmt"
	"log"
	"os"

	docopt "github.com/docopt/docopt-go"

	"github.com/golang-devops/rexec/comms"
)

var (
	Version = "0.0.1"
)

func main() {
	usage := `
		Rexec Client.

		Usage:
		rexec-client exec --server=<server_address> <exe_path> <args>...
		rexec-client start --server=<server_address> <exe_path> <args>...
		rexec-client -h | --help
		rexec-client -v | --version

		Options:
		--server=<server_address>	The address to the server
		-h --help     				Show this screen.
		-v --version  				Show version.
  	`

	arguments, err := docopt.Parse(usage, nil, true, "Rexec Client "+Version, false)
	if err != nil {
		log.Fatal(err)
	}

	serverAddress := arguments["--server"].(string)

	client, err := comms.NewConnectedClient(serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	if arguments["exec"].(bool) == true {
		execExePath := arguments["<exe_path>"].(string)
		execArgs := arguments["<args>"].([]string)
		args := &comms.ExecutorExecuteArgs{Exe: execExePath, Args: execArgs}

		executeReply := &comms.ExecutorExecuteReply{}
		if err := client.Execute(args, executeReply); err != nil {
			log.Fatal(err)
		} else if executeReply.Error != nil {
			log.Fatal(executeReply.Error)
		}

		fmt.Println(fmt.Sprintf("OUT: %s", string(executeReply.Out)))

		os.Exit(0)
	} else if arguments["start"].(bool) == true {
		execExePath := arguments["<exe_path>"].(string)
		execArgs := arguments["<args>"].([]string)
		args := &comms.ExecutorExecuteArgs{Exe: execExePath, Args: execArgs}

		startReply := &comms.ExecutorStartReply{}
		if err := client.Start(args, startReply); err != nil {
			log.Fatal(err)
		} else if startReply.Error != nil {
			log.Fatal(startReply.Error)
		}
		fmt.Println(fmt.Sprintf("Started with PID: %d", startReply.Pid))

		os.Exit(0)
	}
}
