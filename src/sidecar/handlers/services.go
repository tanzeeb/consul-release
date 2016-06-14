package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServicesHandler struct {
	consulURL string
	memberURL string
}

type handlerResponse struct {
	Datacenter string    `json:"datacenter"`
	Services   []service `json:"services"`
}

type service struct {
	Service   string   `json:"service"`
	Addresses []string `json:"addresses"`
}

func NewServicesHandler(consulURL string, memberURL string) ServicesHandler {
	return ServicesHandler{
		consulURL: consulURL,
		memberURL: memberURL,
	}
}

type nodeResponse struct {
	Node node
}

type node struct {
	Address string
}

type selfResponse struct {
	Config config
}

type config struct {
	Datacenter string
}

func (s ServicesHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response, err := http.Get(fmt.Sprintf("%s/v1/agent/self", s.consulURL))
	if err != nil {
		panic(err)
	}

	var selfResp selfResponse
	err = json.NewDecoder(response.Body).Decode(&selfResp)
	if err != nil {
		panic(err)
	}

	response, err = http.Get(fmt.Sprintf("%s/v1/catalog/services", s.consulURL))
	if err != nil {
		panic(err)
	}

	var catalogServicesResp map[string][]string
	err = json.NewDecoder(response.Body).Decode(&catalogServicesResp)
	if err != nil {
		panic(err)
	}

	handlerResponses := []handlerResponse{
		{
			Datacenter: selfResp.Config.Datacenter,
			Services:   []service{},
		},
	}

	for serviceName, serviceNodes := range catalogServicesResp {
		if len(serviceNodes) == 0 {
			continue
		}

		service := service{
			Service:   serviceName,
			Addresses: []string{},
		}

		for _, node := range serviceNodes {
			response, err := http.Get(fmt.Sprintf("%s/v1/catalog/node/%s", s.consulURL, node))
			if err != nil {
				panic(err)
			}

			var nodeResp nodeResponse
			err = json.NewDecoder(response.Body).Decode(&nodeResp)
			if err != nil {
				panic(err)
			}
			service.Addresses = append(service.Addresses, nodeResp.Node.Address)
		}

		handlerResponses[0].Services = append(handlerResponses[0].Services, service)
	}

	if s.memberURL != "" {
		memberResp, err := http.Get(fmt.Sprintf("%s/services", s.memberURL))
		if err != nil {
			panic(err)
		}

		var memberResponse []handlerResponse
		err = json.NewDecoder(memberResp.Body).Decode(&memberResponse)
		if err != nil {
			panic(err)
		}

		handlerResponses = append(handlerResponses, memberResponse[0])
	}

	buf, err := json.Marshal(&handlerResponses)
	if err != nil {
		panic(err)
	}

	w.Write(buf)
}
