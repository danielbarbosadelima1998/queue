package store

import (
	"time"
)

type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
}

type Data struct {
	Payload  string   `json:"payload"`
	Metadata Metadata `json:"metadata"`
}

type ListOptions struct {
	Page    int
	PerPage int
}

type ListItemsResponse struct {
	TotalCount int
	Page       int
	PerPage    int
	Data       []*Data
}

// FIFO
type Store interface {
	Enqueue(queueName, payload string) (index int, err error) // Adiciona um item ao final (Tail)
	Dequeue(queueName string) (data *Data, err error)         // Remove e retorna o primeiro item da fila (Head)
	ListItems(queueName string, options ListOptions) (response *ListItemsResponse, err error)
}
