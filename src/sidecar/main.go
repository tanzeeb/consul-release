package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudfoundry-incubator/consul-release/src/sidecar/handlers"
)

func main() {
	port, consulURL := parseCommandLineFlags()

	serviceHandler := handlers.NewServicesHandler(consulURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/services", func(w http.ResponseWriter, req *http.Request) {
		serviceHandler.ServeHTTP(w, req)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), mux))
}

func parseCommandLineFlags() (string, string) {
	var port string
	var consulURL string

	flag.StringVar(&port, "port", "", "port to use for test consumer server")
	flag.StringVar(&consulURL, "consul-url", "", "url of local consul agent")
	flag.Parse()

	return port, consulURL
}
