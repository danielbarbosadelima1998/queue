package store

import (
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
)

type MemoryStore struct {
	// { queueName: []string }
	Data map[string][]string // @TODO: usar uma lista duplamente encadeada *list.List ?
	mu   sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Data: make(map[string][]string, 0),
	}
}

func (s *MemoryStore) Enqueue(queueName string, payload string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[queueName] = append(s.Data[queueName], payload)

	return len(s.Data[queueName]), nil
}

func (s *MemoryStore) Dequeue(queueName string) (data *Data, err error) {
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
		return nil, fmt.Errorf("failed to dequeue on parser json to struct %s ", err.Error())
	}

	s.Data[queueName] = s.Data[queueName][1:]

	return data, nil
}

func paginate(totalItems, page, perPage int) (from, to int) {
	if perPage <= 0 {
		perPage = 10 // valor padrão
	}
	if page < 1 {
		page = 1
	}

	from = (page - 1) * perPage
	if from > totalItems {
		from = totalItems
	}

	to = from + perPage
	if to > totalItems {
		to = totalItems
	}

	return from, to
}

func (s *MemoryStore) ListItems(queueName string, options ListOptions) (response *ListItemsResponse, err error) {
	totalCount := len(s.Data[queueName])
	from, to := paginate(totalCount, options.Page, options.PerPage)

	fmt.Println("Listing Items on queueName: ", queueName,
		"\npage: ", options.Page,
		"\nperPage: ", options.PerPage,
		"\nfrom: ", from,
		"\nto: ", to,
		"\ntotalCount: ", totalCount,
	)

	items := s.Data[queueName][from:to]

	data := make([]*Data, len(items))

	for key, value := range items {
		if value == "" {
			return nil, fmt.Errorf("failed to dequeue, data on index 0 is empty")
		}

		err = sonic.UnmarshalString(value, &data[key])
		if err != nil {
			return nil, fmt.Errorf("failed to dequeue on parser json to struct %s ", err.Error())
		}
	}

	return &ListItemsResponse{
		Data:       data,
		TotalCount: totalCount,
		Page:       options.Page,
		PerPage:    options.PerPage,
	}, nil
}
