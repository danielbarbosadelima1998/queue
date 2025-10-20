package queue

import (
	"fmt"
	"queue/store"
	"testing"
)

func ValidateDequeueResponse(t *testing.T, payloadExpected string, data *store.Data, err error) {
	if err != nil {
		t.Errorf("dequeue should succeed without errors, returned: %v", err)
	}

	if data == nil {
		t.Errorf("the data should exists, returned: %v", data)
	}

	if data.Metadata.Timestamp.IsZero() {
		t.Errorf("the data should contain a valid timestamp Metadata, returned: %v", data.Metadata.Timestamp)
	}

	if data.Payload != payloadExpected {
		t.Errorf("the data should contain the same enqueued payload, expected: %s returned: %s", payloadExpected, data.Payload)
	}
}

func ValidateEnqueueResponse(t *testing.T, indexExpected int, index int, err error) {
	if err != nil {
		t.Errorf("enqueueing should succeed without errors, returned: %v", err)
	}

	if index != indexExpected {
		t.Errorf("enqueueing index should be %d, returned: %d", indexExpected, index)
	}
}

func TestSequentialQueueBasic(t *testing.T) {
	queue := NewSequentialQueue(QueueConfig{
		Store: store.NewMemoryStore(),
		Name:  "messages",
	})

	data, err := queue.Dequeue()

	if err != nil {
		t.Errorf("dequeue empty queue should succeed without error, returned: %v", err)
	}

	if data != nil {
		t.Errorf("dequeue empty queue should return nil data, returned: %v", data)
	}

	payload := "Message 1"

	index, err := queue.Enqueue(payload)
	ValidateEnqueueResponse(t, 0, index, err)

	data, err = queue.Dequeue()

	ValidateDequeueResponse(t, payload, data, err)
}

func TestSequentialQueueManyItens(t *testing.T) {
	queue := NewSequentialQueue(QueueConfig{
		Store: store.NewMemoryStore(),
		Name:  "messages",
	})

	for i := 0; i < 1000; i++ {
		payload := fmt.Sprintf("Message %d", i)
		index, err := queue.Enqueue(payload)
		ValidateEnqueueResponse(t, i, index, err)
	}

	for i := 0; i < 1000; i++ {
		data, err := queue.Dequeue()
		payloadExpected := fmt.Sprintf("Message %d", i)
		ValidateDequeueResponse(t, payloadExpected, data, err)
	}
}
