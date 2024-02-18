package node

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/voIatiIe/tidewave/cmd/entities"
)

func (n *Node) requestHandler(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusOK)

	if r.Method != "POST" {
		return
	}

	var request entities.Request
	json.NewDecoder(r.Body).Decode(&request)

	if n.requestLog.Exists(request.ID) {
		return
	}
	n.requestLog.Put(request.ID, struct{}{})

	if value, ok := n.data.Get(request.ResourceId); ok {
		message, _ := json.Marshal(entities.Response{
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

	var response entities.Response
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

	var request entities.GetRequest
	json.NewDecoder(r.Body).Decode(&request)

	resource, err := n.Get(request.ResourceId)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	message, _ := json.Marshal(entities.GetResponse{
		ResourceId: request.ResourceId,
		Resource:   resource,
	})

	w.Write(message)
}

func (n *Node) putHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var request entities.PutRequest
	json.NewDecoder(r.Body).Decode(&request)

	n.data.Put(request.ResourceId, request.Resource)
}

func (n *Node) deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var request entities.DeleteRequest
	json.NewDecoder(r.Body).Decode(&request)

	n.data.Delete(request.ResourceId)
}
