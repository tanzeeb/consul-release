package main_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sidecar", func() {
	var (
		session            *gexec.Session
		port               string
		consulServer       *httptest.Server
		otherSidecarServer *httptest.Server
	)

	Context("services", func() {
		BeforeEach(func() {
			consulServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/v1/agent/self":
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"Config": {
							"Datacenter": "dc1"
						}
					}`))
					return
				case "/v1/catalog/services":
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"consul":[],
						"my-dc1-service": ["node-1"],
						"my-other-dc1-service": ["node-1"]
					}`))
					return
				case "/v1/catalog/node/node-1":
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"Node": {
							"Address": "10.0.0.2"
						}
					}`))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))

			otherSidecarServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte(`[
					{
						"datacenter": "dc2",
						"services": [
							{"service": "my-dc2-service","addresses": ["10.0.1.1","10.0.1.2","10.0.1.3"]},
							{"service": "my-other-dc2-service","addresses": ["10.0.1.1","10.0.1.3"]}
						]
					}
				]`))
			}))

			var err error
			port, err = openPort()
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(pathToSidecar, "--port", port, "--consul-url", consulServer.URL,
				"--member", otherSidecarServer.URL)

			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			waitForServerToStart(port)
		})

		AfterEach(func() {
			session.Terminate().Wait()
		})

		It("returns services which have been scraped from consul", func() {
			status, responseBodyString, err := makeRequest("GET", fmt.Sprintf("http://localhost:%s/services", port), "")
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))
			Expect(responseBodyString).To(MatchJSON(`[
				{
					"datacenter": "dc1",
					"services": [
						{"service": "my-dc1-service","addresses": ["10.0.0.2"]},
						{"service": "my-other-dc1-service","addresses": ["10.0.0.2"]}
					]
				},
				{
					"datacenter": "dc2",
					"services": [
						{"service": "my-dc2-service","addresses": ["10.0.1.1","10.0.1.2","10.0.1.3"]},
						{"service": "my-other-dc2-service","addresses": ["10.0.1.1","10.0.1.3"]}
					]
				}
			]`))
		})
	})
})

func waitForServerToStart(port string) {
	timer := time.After(0 * time.Second)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			panic("Failed to boot!")
		case <-timer:
			_, err := http.Get("http://localhost:" + port + "/services")
			if err == nil {
				return
			}

			timer = time.After(1 * time.Second)
		}
	}
}

func openPort() (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}

	return port, nil
}

func makeRequest(method string, url string, body string) (int, string, error) {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, "", err
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, "", err
	}

	return response.StatusCode, string(responseBody), nil
}
