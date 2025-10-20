package queue

import "queue/store"

// FIFO
type Queue interface {
	Enqueue(payload string) (index int, err error) // Adiciona um item ao final (Tail)
	Dequeue() (data *store.Data, err error)        // Remove e retorna o primeiro item da fila (Head)
	ListItems(options store.ListOptions) (ListItems *store.ListItemsResponse, err error)
}

type QueueConfig struct {
	Store store.Store
	Name  string
}
