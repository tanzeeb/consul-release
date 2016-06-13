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

	if nodeName == "node-2" {
		payload["Services"] = map[string]interface{}{
			"my-service": map[string]string{
				"ID":      "my-service",
				"Service": "my-service",
			},
		}
	} else {
		payload["Services"] = map[string]interface{}{
			"my-service": map[string]string{
				"ID":      "my-service",
				"Service": "my-service",
			},
			"my-other-service": map[string]string{
				"ID":      "my-other-service",
				"Service": "my-other-service",
			},
		}
	}

	return json.Marshal(&payload)
}

var _ = Describe("services", func() {
	It("scrapes the consul agent for service definitions", func() {
		consulURL := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
			}
			w.WriteHeader(http.StatusTeapot)
		}))

		request, err := http.NewRequest("GET", "/services", nil)
		Expect(err).NotTo(HaveOccurred())

		handler := handlers.NewServicesHandler(consulURL.URL)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.Bytes()).To(MatchJSON(`{
			"services": [
				{"service": "my-service","addresses": ["10.0.0.1","10.0.0.2","10.0.0.3"]},
				{"service": "my-other-service","addresses": ["10.0.0.1","10.0.0.3"]}
			]
		}`))
	})
})
