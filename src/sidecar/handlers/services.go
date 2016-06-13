package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServicesHandler struct {
	consulURL string
}

type service struct {
	Service   string   `json:"service"`
	Addresses []string `json:"addresses"`
}

type services struct {
	Services []service `json:"services"`
}

func NewServicesHandler(consulURL string) ServicesHandler {
	return ServicesHandler{
		consulURL: consulURL,
	}
}

type nodeResponse struct {
	Node node
}

type node struct {
	Address string
}

func (s ServicesHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response, err := http.Get(fmt.Sprintf("%s/v1/catalog/services", s.consulURL))
	if err != nil {
		panic(err)
	}

	var catalogServicesResp map[string][]string
	err = json.NewDecoder(response.Body).Decode(&catalogServicesResp)
	if err != nil {
		panic(err)
	}

	services := services{
		Services: []service{},
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

		services.Services = append(services.Services, service)
	}

	buf, err := json.Marshal(&services)
	if err != nil {
		panic(err)
	}

	w.Write(buf)
}
