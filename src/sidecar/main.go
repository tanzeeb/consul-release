package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudfoundry-incubator/consul-release/src/sidecar/handlers"
)

func main() {
	port, consulURL, sidecarMemberURL := parseCommandLineFlags()

	serviceHandler := handlers.NewServicesHandler(consulURL, sidecarMemberURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/services", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%+v\n", req)
		serviceHandler.ServeHTTP(w, req)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), mux))
}

func parseCommandLineFlags() (string, string, string) {
	var port string
	var consulURL string
	var sidecarMemberURL string

	flag.StringVar(&port, "port", "", "port to use for test consumer server")
	flag.StringVar(&consulURL, "consul-url", "", "url of local consul agent")
	flag.StringVar(&sidecarMemberURL, "member", "", "url of a remote sidcar")
	flag.Parse()

	return port, consulURL, sidecarMemberURL
}
