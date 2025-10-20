package queue

import (
	"queue/store"
	"testing"
)

func BenchmarkEnqueue(b *testing.B) {
	queue := NewSequentialQueue(QueueConfig{
		Store: store.NewMemoryStore(),
		Name:  "messages",
	})

	for b.Loop() {
		_, _ = queue.Enqueue("Message")
	}
}

func BenchmarkDequeue(b *testing.B) {
	store := store.NewMemoryStore()
	queue := NewSequentialQueue(QueueConfig{
		Store: store,
		Name:  "messages",
	})

	for b.Loop() {
		_, _ = queue.Enqueue("Message")
		_, _ = queue.Dequeue()
	}
}
