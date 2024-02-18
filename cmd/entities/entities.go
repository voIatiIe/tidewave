package entities

import (
	"errors"
	"strings"

	"github.com/voIatiIe/tidewave/cmd/tool"
)

type RequestLog struct {
	tool.SafeHashMap[string, struct{}]
}
type AwaitQueue struct {
	tool.SafeHashMap[string, chan string]
}
type Data struct {
	tool.SafeHashMap[string, string]
}

func (r *Data) String() string {
	return "Routing table"
}

func (r *Data) Set(value string) error {
	values := strings.SplitN(value, ":", 2)
	if len(values) != 2 {
		return errors.New("wrong argument format")
	}

	r.Put(values[0], values[1])

	return nil
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

type PutRequest struct {
	ResourceId string `json:"resource_id"`
	Resource   string `json:"resource"`
}

type DeleteRequest struct {
	ResourceId string `json:"resource_id"`
}

func NewRequestLog() RequestLog {
	return RequestLog{
		SafeHashMap: tool.NewSafeHashMap[string, struct{}](),
	}
}

func NewAwaitQueue() AwaitQueue {
	return AwaitQueue{
		SafeHashMap: tool.NewSafeHashMap[string, chan string](),
	}
}

func NewData() Data {
	return Data{
		SafeHashMap: tool.NewSafeHashMap[string, string](),
	}
}
