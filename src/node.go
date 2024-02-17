package src

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

	"github.com/google/uuid"
)

type RequestLog struct {
	SafeHashMap[string, struct{}]
}
type AwaitQueue struct {
	SafeHashMap[string, chan string]
}

type Request struct {
	ID         string `json:"id"`
	ResourceId string `json:"resource_id"`
	Origin     string `json:"origin"`
}

type Response struct {
	Request
	Body string `json:"body"`
}

type GetRequest struct {
	ResourceId string `json:"resource_id"`
}

type GetResponse struct {
	ResourceId string `json:"resource_id"`
	Resource   string `json:"resource"`
}

type Node struct {
	id           uuid.UUID
	addr         string
	data         map[string]string
	routingTable []string
	requestLog   RequestLog
	queue        AwaitQueue
	server       http.Server
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewNode(config Config) Node {
	ctx, cancel := context.WithCancel(context.Background())

	return Node{
		id:           uuid.New(),
		addr:         config.addr,
		data:         config.data,
		routingTable: config.routingTable,
		requestLog:   NewRequestLog(),
		queue:        NewAwaitQueue(),
		server:       http.Server{Addr: config.addr},
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (n *Node) Start() {
	http.HandleFunc("/request", n.requestHandler)
	http.HandleFunc("/response", n.responseHandler)
	http.HandleFunc("/get", n.getHandler)

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

func (n *Node) requestHandler(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusOK)

	if r.Method != "POST" {
		return
	}

	var request Request
	json.NewDecoder(r.Body).Decode(&request)

	if n.requestLog.Exists(request.ID) {
		fmt.Println("Reject")
		return
	}
	fmt.Println("Request")
	n.requestLog.Put(request.ID, struct{}{})

	if value, ok := n.data[request.ResourceId]; ok {
		message, _ := json.Marshal(Response{
			Request: request,
			Body:    value,
		})

		request, _ := http.NewRequest("POST", "http://localhost"+request.Origin+"/response", bytes.NewBuffer(message))

		client := http.Client{}
		client.Do(request)

		return
	}

	for _, addr := range n.routingTable {
		message, _ := json.Marshal(request)
		request, _ := http.NewRequest("POST", "http://localhost"+addr+"/request", bytes.NewBuffer(message))

		client := http.Client{}
		client.Do(request)
	}
}

func (n *Node) responseHandler(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusOK)

	if r.Method != "POST" {
		return
	}

	var response Response
	json.NewDecoder(r.Body).Decode(&response)

	if resultChannel, ok := n.queue.Get(response.ResourceId); ok {
		n.queue.Delete(response.ResourceId)

		resultChannel <- response.Body
	}
}

func (n *Node) getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var request GetResponse
	json.NewDecoder(r.Body).Decode(&request)

	resource, err := n.Get(request.ResourceId)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	message, _ := json.Marshal(GetResponse{
		ResourceId: request.ResourceId,
		Resource:   resource,
	})

	w.Write(message)
}

func (n *Node) Get(resourceId string) (string, error) {
	value, ok := n.data[resourceId]

	if ok {
		return value, nil
	}

	resultChan := make(chan string)
	n.queue.Put(resourceId, resultChan)

	request := Request{
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

		fmt.Println("Timeout")

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

func NewRequestLog() RequestLog {
	return RequestLog{
		SafeHashMap[string, struct{}]{
			map_: make(map[string]struct{}),
		},
	}
}

func NewAwaitQueue() AwaitQueue {
	return AwaitQueue{
		SafeHashMap[string, chan string]{
			map_: make(map[string]chan string),
		},
	}
}
