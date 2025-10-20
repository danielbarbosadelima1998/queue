package utils

import (
	"fmt"
	"sync"
)

type Worker struct {
	name        string
	concurrency int
	jobs        chan (Job)
	wg          sync.WaitGroup
}

type Job func(workerNumber int) (response any, err error)

func NewWorker(name string, concurrency int) *Worker {
	return &Worker{
		name:        name,
		concurrency: concurrency,
		jobs:        make(chan Job, concurrency),
	}
}

func (w *Worker) worker(workerNumber int) {
	defer w.wg.Done()

	for job := range w.jobs {
		_, _ = job(workerNumber)

		// log adiciona MUITO overhead desnecessário
		// fmt.Printf("Running job on worker \"%s\", response: %v error: %v\n", workerNumber, response, err)
	}
}

func (w *Worker) Start() {
	fmt.Printf("Starting worker \"%s\" concurrency %d\n", w.name, w.concurrency)

	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		// Inicia uma goroutine que fica consumindo jobs até terminar
		go w.worker(i)
	}
}
func (w *Worker) RunJob(job Job) (response any, err error) {
	// Adiciona o Job para ser consumindo pelos workers
	w.jobs <- job

	return response, err
}

func (w *Worker) CloseWorker() {
	close(w.jobs)
	w.wg.Wait()
	fmt.Printf("Closed worker \"%s\"\n", w.name)
}
