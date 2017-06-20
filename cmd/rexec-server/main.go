package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strings"

	"github.com/apex/log"

	_ "github.com/golang-devops/rexec/comms"
	"github.com/golang-devops/rexec/logging"
)

var (
	Version string = "0.0.2"
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

func main() {
	logger := logging.Logger()
	logger.Info(fmt.Sprintf("Version %s", Version))

	flag.Parse()
	if *address == "" {
		logger.Fatal("address flag is missing")
	}

	printMyIPs(logger)

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
