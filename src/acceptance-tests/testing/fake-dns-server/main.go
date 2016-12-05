package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/cloudfoundry-incubator/consul-release/src/acceptance-tests/testing/fake-dns-server/dnsserver"
)

func main() {
	log.Println("Starting dns server...")
	server := dnsserver.NewServer()
	server.Start()
	log.Println("Started dns server")

	log.Printf("Registering %s %s\n", "my-fake-server.fake.local", "10.2.3.4")
	server.RegisterARecord("my-fake-server.fake.local", net.ParseIP("10.2.3.4"))
	server.RegisterAAAARecord("my-fake-server.fake.local", net.ParseIP("10.2.3.4"))

	UDP_TRUNCATION_THRESHOLD := 4
	for i := 0; i < UDP_TRUNCATION_THRESHOLD; i++ {
		server.RegisterARecord("large-dns-response", net.ParseIP(makeIP()))
	}

	select {}
}

func makeIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}
