package queue

import (
	"fmt"
	"log"
	"queue/store"
	"time"

	"github.com/bytedance/sonic"
)

type SequentialQueue struct {
	store store.Store
	name  string
}

func NewSequentialQueue(config QueueConfig) Queue {
	if config.Name == "" {
		log.Panic("config name is required to create sequential queue")
	}

	if config.Store == nil {
		log.Panic("config store is required to create sequential queue")
	}

	return &SequentialQueue{
		name:  config.Name,
		store: config.Store,
	}
}

func (q *SequentialQueue) Enqueue(payload string) (int, error) {
	if payload == "" {
		return 0, fmt.Errorf("failed to enqueue empty value")
	}

	dataString, err := sonic.MarshalString(store.Data{
		Payload: payload,
		Metadata: store.Metadata{
			Timestamp: time.Now(),
		},
	})

	if err != nil {
		return 0, fmt.Errorf("failed to encode data to json %s", err.Error())
	}

	return q.store.Enqueue(q.name, dataString)
}

func (q *SequentialQueue) Dequeue() (*store.Data, error) {
	return q.store.Dequeue(q.name)
}

func (q *SequentialQueue) ListItems(options store.ListOptions) (*store.ListItemsResponse, error) {
	return q.store.ListItems(q.name, store.ListOptions{
		Page:    options.Page,
		PerPage: options.PerPage,
	})
}
