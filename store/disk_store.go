package store

import (
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
)

type DiskStore struct {
	// { queueName: []string
	Data map[string][]string
	mu   sync.Mutex
}

func NewDiskStore() *DiskStore {
	return &DiskStore{
		Data: make(map[string][]string, 0),
	}
}

func (s *DiskStore) Enqueue(queueName string, payload string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[queueName] = append(s.Data[queueName], payload)

	return len(s.Data[queueName]) - 1, nil
}

func (s *DiskStore) Dequeue(queueName string) (data *Data, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueLength := len(s.Data[queueName])

	if queueLength == 0 {
		return nil, nil // no return error
	}

	value := s.Data[queueName][0]

	if value == "" {
		return nil, fmt.Errorf("failed to dequeue, data on index 0 is empty")
	}

	err = sonic.UnmarshalString(value, &data)

	if err != nil {
		return nil, fmt.Errorf("failed to dequeue, parser of json to struct data failed %s ", err.Error())
	}

	s.Data[queueName] = s.Data[queueName][1:]

	return data, nil
}
