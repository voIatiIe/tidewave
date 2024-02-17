package src

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	addr         string
	data         map[string]string
	routingTable []string
}

type RoutingTable []string
type Data map[string]string

func (r *RoutingTable) String() string {
	return "Routing table"
}

func (r *RoutingTable) Set(value string) error {
	*r = append(*r, fmt.Sprintf(":%s", value))

	return nil
}

func (r *Data) String() string {
	return "Routing table"
}

func (r *Data) Set(value string) error {
	values := strings.SplitN(value, ":", 2)

	(*r)[values[0]] = values[1]

	return nil
}

func ParseCommandLine() Config {
	var port int
	var routingTable RoutingTable
	var data Data = make(map[string]string)

	flag.IntVar(&port, "p", 9070, "Node port")
	flag.Var(&routingTable, "r", "Node routing table")
	flag.Var(&data, "d", "Node data")

	flag.Parse()

	return Config{
		addr:         fmt.Sprintf(":%d", port),
		routingTable: routingTable,
		data:         data,
	}
}
