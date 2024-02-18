package config

import (
	"flag"
	"fmt"

	"github.com/voIatiIe/tidewave/cmd/entities"
)

type Config struct {
	Addr         string
	Data         entities.Data
	RoutingTable []string
}

type RoutingTable []string

func (r *RoutingTable) String() string {
	return "Routing table"
}

func (r *RoutingTable) Set(value string) error {
	*r = append(*r, fmt.Sprintf(":%s", value))

	return nil
}

func ParseCommandLine() Config {
	var port int
	var routingTable RoutingTable
	var data entities.Data = entities.NewData()

	flag.IntVar(&port, "p", 9070, "Node port")
	flag.Var(&routingTable, "r", "Node routing table")
	flag.Var(&data, "d", "Node data")

	flag.Parse()

	return Config{
		Addr:         fmt.Sprintf(":%d", port),
		RoutingTable: routingTable,
		Data:         data,
	}
}
