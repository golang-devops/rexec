package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var (
	address = flag.String("address", "127.0.0.1:50500", "The address to listen on")
)

func main() {
	flag.Parse()
	if *address == "" {
		log.Fatal("address flag is missing")
	}

	prefixHandlers := map[string]func(remainingURL string, w http.ResponseWriter, r *http.Request){
		"/exec-wait": func(remainingURL string, w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				writeResponseError(w, http.StatusNotFound, "Only POST method allowed")
				return
			}
			defer r.Body.Close()

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				msg := fmt.Sprintf("Unable to read body, error: %s", err.Error())
				writeResponseError(w, http.StatusBadRequest, msg)
				return
			}

			execParts := strings.Split(strings.TrimSpace(string(body)), "\n")
			for i, p := range execParts {
				execParts[i] = strings.TrimSpace(p)
			}

			cmd := exec.Command(execParts[0], execParts[1:]...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				msg := fmt.Sprintf("Command execution failed. Error: %s. Output: %s", err.Error(), string(out))
				writeResponseError(w, http.StatusInternalServerError, msg)
				return
			}

			writeResponseSuccessMessage(w, fmt.Sprintf("Done, output: %s", string(out)))
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		origPath := r.URL.Path
		lowerPath := strings.ToLower(origPath)

		gotHandled := false
		for prefix, fn := range prefixHandlers {
			fmt.Println(fmt.Sprintf("%s|%s", r.Method, origPath))

			if !strings.HasPrefix(lowerPath, prefix) {
				continue
			}

			remainingURL := origPath[len(prefix):]
			fn(remainingURL, w, r)
			gotHandled = true
		}

		if !gotHandled {
			log.Println("URL unhandled: " + origPath)
		}
	})

	fmt.Println("Listening on " + *address)
	http.ListenAndServe(*address, nil)
}
