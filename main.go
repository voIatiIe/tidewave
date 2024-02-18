package main

import (
	"github.com/voIatiIe/tidewave/cmd/config"
	"github.com/voIatiIe/tidewave/cmd/node"
)

func main() {
	config := config.ParseCommandLine()

	node := node.NewNode(&config)
	node.Start()
	node.AwaitShutdown()
}
