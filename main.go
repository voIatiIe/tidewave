package main

import (
	"tide/src"
)


func main() {
	config := src.ParseCommandLine()

	node := src.NewNode(config)
	node.Start()
	node.AwaitShutdown()
}
