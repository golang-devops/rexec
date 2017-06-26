package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/apex/log"
	"github.com/docker/libchan"
	"github.com/docker/libchan/spdy"

	"github.com/golang-devops/rexec/comms"
	"github.com/golang-devops/rexec/logging"
)

var (
	// address = flag.String("address", "127.0.0.1:50505", "The address to listen on")
	address = flag.String("address", "0.0.0.0:50505", "The address to listen on")
)

func printMyIPs(logger *log.Entry) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.WithError(err).Warn("Failed to list network Interfaces")
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.WithError(err).Warnf("Failed to get network address of %s", i.Name)
			continue
		}

		lowerName := strings.ToLower(i.Name)
		ignoreIfContains := []string{
			"Virtual",
			"Docker",
		}

		shouldBeIgnored := false
		for _, toIgnore := range ignoreIfContains {
			if strings.Contains(lowerName, strings.ToLower(toIgnore)) {
				shouldBeIgnored = true
				break
			}
		}
		if shouldBeIgnored {
			logger.Warnf("Skipping obtaining IP of %s (contains one of %s)", i.Name, strings.Join(ignoreIfContains, ", "))
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					logger.Infof("IP %s: %s", i.Name, ipnet.IP.String())
				}
			}
		}
	}
}

// RemoteCommand is the received command parameters to execute locally and return
type RemoteCommand struct {
	Cmd        string
	Args       []string
	Stdin      io.Reader
	Stdout     io.WriteCloser
	Stderr     io.WriteCloser
	StatusChan libchan.Sender
}

// CommandResponse is the response struct to return to the client
type CommandResponse struct {
	Status int
}

func main() {
	logger := logging.Logger()
	logger.Info(fmt.Sprintf("Version %s", comms.Version))

	flag.Parse()
	if *address == "" {
		logger.Fatal("address flag is missing")
	}

	printMyIPs(logger)

	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")

	var listener net.Listener
	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			logger.WithError(err).Fatal("Failed to LoadX509KeyPair")
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{tlsCert},
		}

		logger.Info(fmt.Sprintf("TCP (with TLS) server listening on %s", *address))
		listener, err = tls.Listen("tcp", *address, tlsConfig)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to Listen on address %s", *address)
		}
	} else {
		var err error
		logger.Info(fmt.Sprintf("TCP (non-TLS) server listening on %s", *address))
		listener, err = net.Listen("tcp", *address)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to Listen on address %s", *address)
		}
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			logger.WithError(err).Error("Failed to accept Client connection")
			break
		}
		p, err := spdy.NewSpdyStreamProvider(c, true)
		if err != nil {
			logger.WithError(err).Error("Failed to create Spdy Stream Provider")
			break
		}
		t := spdy.NewTransport(p)

		go func() {
			for {
				receiver, err := t.WaitReceiveChannel()
				if err != nil {
					logger.WithError(err).Error("Failed to wait for Receive Channel")
					break
				}

				go func() {
					for {
						command := &RemoteCommand{}
						err := receiver.Receive(command)
						if err != nil {
							logger.WithError(err).Error("Failed to Receive command")
							break
						}

						cmd := exec.Command(command.Cmd, command.Args...)
						cmd.Stdout = command.Stdout
						cmd.Stderr = command.Stderr

						stdin, err := cmd.StdinPipe()
						if err != nil {
							logger.WithError(err).Error("Failed to get command StdinPipe")
							break
						}
						go func() {
							io.Copy(stdin, command.Stdin)
							stdin.Close()
						}()

						runErr := cmd.Run()
						command.Stdout.Close()
						command.Stderr.Close()
						returnResult := &CommandResponse{}
						if runErr != nil {
							if exiterr, ok := runErr.(*exec.ExitError); ok {
								returnResult.Status = exiterr.Sys().(syscall.WaitStatus).ExitStatus()
							} else {
								logger.WithError(runErr).Error("Command Run failed")
								returnResult.Status = 10
							}
						}

						err = command.StatusChan.Send(returnResult)
						if err != nil {
							logger.WithError(err).Error("Failed to send result to Client")
						}
					}
				}()
			}
		}()
	}
}
