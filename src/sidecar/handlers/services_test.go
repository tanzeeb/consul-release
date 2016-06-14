package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cloudfoundry-incubator/consul-release/src/sidecar/handlers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func buildJSONForNode(nodeName string) ([]byte, error) {
	ipAddress := fmt.Sprintf("10.0.0.%s", strings.Split(nodeName, "-")[1])

	payload := map[string]interface{}{
		"Node": map[string]interface{}{
			"Node":    nodeName,
			"Address": ipAddress,
			"TaggedAddresses": map[string]string{
				"wan": ipAddress,
			},
		},
	}

	return json.Marshal(&payload)
}

var _ = Describe("services", func() {
	var (
		consulServer        *httptest.Server
		sidecarMemberServer *httptest.Server
	)

	BeforeEach(func() {
		consulServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch {
			case req.URL.Path == "/v1/catalog/services":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"consul":[],
					"my-service": ["node-1", "node-2", "node-3"],
					"my-other-service": ["node-1", "node-3"]
				}`))
				return
			case strings.HasPrefix(req.URL.Path, "/v1/catalog/node"):
				parts := strings.SplitAfter(req.URL.Path, "/v1/catalog/node/")
				payload, err := buildJSONForNode(parts[1])
				Expect(err).NotTo(HaveOccurred())

				w.WriteHeader(http.StatusOK)
				w.Write(payload)
				return
			case req.URL.Path == "/v1/agent/self":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"Config": {
						"Datacenter": "dc1"
					}
				}`))
				return
			}
			w.WriteHeader(http.StatusTeapot)
		}))

		sidecarMemberServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/services":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"datacenter": "dc2",
						"services": [
							{"service": "my-dc2-service","addresses": ["10.0.1.1","10.0.1.2","10.0.1.3"]},
							{"service": "my-other-dc2-service","addresses": ["10.0.1.1","10.0.1.3"]}
						]
					}
				]`))
				return
			}
			w.WriteHeader(http.StatusTeapot)
		}))
	})

	It("scrapes the consul agent for service definitions", func() {
		request, err := http.NewRequest("GET", "/services", nil)
		Expect(err).NotTo(HaveOccurred())

		handler := handlers.NewServicesHandler(consulServer.URL, "")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.Bytes()).To(MatchJSON(`[
			{
				"datacenter": "dc1",
				"services": [
					{"service": "my-service","addresses": ["10.0.0.1","10.0.0.2","10.0.0.3"]},
					{"service": "my-other-service","addresses": ["10.0.0.1","10.0.0.3"]}
				]
			}
		]`))
	})

	It("appends member services to the returned service definitions", func() {
		request, err := http.NewRequest("GET", "/services", nil)
		Expect(err).NotTo(HaveOccurred())

		handler := handlers.NewServicesHandler(consulServer.URL, sidecarMemberServer.URL)

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.Bytes()).To(MatchJSON(`[
			{
				"datacenter": "dc1",
				"services": [
					{"service": "my-service","addresses": ["10.0.0.1","10.0.0.2","10.0.0.3"]},
					{"service": "my-other-service","addresses": ["10.0.0.1","10.0.0.3"]}
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
