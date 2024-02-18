package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voIatiIe/tidewave/cmd/config"
	"github.com/voIatiIe/tidewave/cmd/entities"

	"github.com/google/uuid"
)

type Node struct {
	id           uuid.UUID
	addr         string
	data         entities.Data
	routingTable []string
	requestLog   entities.RequestLog
	queue        entities.AwaitQueue
	server       http.Server
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewNode(config *config.Config) Node {
	ctx, cancel := context.WithCancel(context.Background())

	return Node{
		id:           uuid.New(),
		addr:         config.Addr,
		data:         config.Data,
		routingTable: config.RoutingTable,
		requestLog:   entities.NewRequestLog(),
		queue:        entities.NewAwaitQueue(),
		server:       http.Server{Addr: config.Addr},
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (n *Node) Start() {
	http.HandleFunc("/request", n.requestHandler)
	http.HandleFunc("/response", n.responseHandler)
	http.HandleFunc("/get", n.getHandler)
	http.HandleFunc("/put", n.putHandler)
	http.HandleFunc("/delete", n.deleteHandler)

	go func() {
		if err := n.server.ListenAndServe(); err != nil {
			fmt.Println("Node:", err)

			n.cancel()
		}
	}()

	fmt.Printf("Starting node at %s\n", n.addr)
}

func (n *Node) Shutdown() {
	n.server.Shutdown(n.ctx)
}

func (n *Node) Get(resourceId string) (string, error) {
	value, ok := n.data.Get(resourceId)

	if ok {
		return value, nil
	}

	resultChan := make(chan string)
	n.queue.Put(resourceId, resultChan)

	request := entities.Request{
		ID:         uuid.NewString(),
		ResourceId: resourceId,
		Origin:     n.addr,
	}
	n.requestLog.Put(request.ID, struct{}{})

	go func() {

		for _, addr := range n.routingTable {
			message, _ := json.Marshal(request)
			url := "http://localhost" + addr + "/request"

			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(message))

			client := http.Client{}
			client.Do(req)
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil

	case <-time.After(2 * time.Second):
		n.queue.Delete(resourceId)
		close(resultChan)

		return "", errors.New("timout error")
	}
}

func (n *Node) AwaitShutdown() {
	ret := make(chan struct{})
	go func() {
		exit := make(chan os.Signal, 2)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

		select {
		case <-exit:
			fmt.Println("\nShutting down node...")
			n.Shutdown()

			n.cancel()
			ret <- struct{}{}

		case <-n.ctx.Done():
			return
		}
	}()

	select {
	case <-ret:
		return
	case <-n.ctx.Done():
		return
	}
}
