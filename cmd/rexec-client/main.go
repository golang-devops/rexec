package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"

	"github.com/docker/libchan"
	"github.com/docker/libchan/spdy"
	docopt "github.com/docopt/docopt-go"

	"github.com/golang-devops/rexec/comms"
)

// RemoteCommand is the run parameters to be executed remotely
type RemoteCommand struct {
	Cmd        string
	Args       []string
	Stdin      io.Writer
	Stdout     io.Reader
	Stderr     io.Reader
	StatusChan libchan.Sender
}

// CommandResponse is the returned response object from the remote execution
type CommandResponse struct {
	Status int
}

func main() {
	usage := `
		Rexec Client.

		Usage:
		rexec-client [--use-tls] --server=<server_address> <exe_path> <args>...
		rexec-client -h | --help
		rexec-client -v | --version

		Options:
		--use-tls                   Whether to use TLS for server communication.
		--server=<server_address>	The address to the server.
		-h --help     				Show this screen.
		-v --version  				Show version.
  	`

	arguments, err := docopt.Parse(usage, nil, true, "Rexec Client "+comms.Version, false)
	if err != nil {
		log.Fatal(err)
	}

	serverAddress := arguments["--server"].(string)

	var client net.Conn
	var dialErr error
	if os.Getenv("USE_TLS") != "" {
		client, dialErr = tls.Dial("tcp", serverAddress, &tls.Config{InsecureSkipVerify: true})
	} else {
		client, dialErr = net.Dial("tcp", serverAddress)
	}
	if dialErr != nil {
		log.Fatal(dialErr)
	}

	streamProvider, err := spdy.NewSpdyStreamProvider(client, false)
	if err != nil {
		log.Fatal(err)
	}

	transport := spdy.NewTransport(streamProvider)
	sender, err := transport.NewSendChannel()
	if err != nil {
		log.Fatal(err)
	}

	receiver, remoteSender := libchan.Pipe()

	command := &RemoteCommand{
		Cmd:        arguments["<exe_path>"].(string),
		Args:       arguments["<args>"].([]string),
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		StatusChan: remoteSender,
	}

	err = sender.Send(command)
	if err != nil {
		log.Fatal(err)
	}

	response := &CommandResponse{}
	err = receiver.Receive(response)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(response.Status)
}
