package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	totalRequests   = 100000 // Total de requisições
	concurrentLimit = 1000   // Número máximo de workers simultâneos
	url             = "http://localhost:7500/api/v1/queue/messages/enqueue"
)

var (
	client *http.Client
)

func init() {
	// Cliente HTTP global e reutilizável com conexões otimizadas
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10000,
			MaxIdleConnsPerHost: 10000,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		payload := fmt.Sprintf(`{"teste": "Message %d"}`, j)
		req, err := http.NewRequest("POST", url, io.NopCloser(strings.NewReader(payload)))
		if err != nil {
			fmt.Printf("Worker %d: erro ao criar request: %v\n", id, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Worker %d: erro na requisição: %v\n", id, err)
			continue
		}
		io.Copy(io.Discard, resp.Body) // Descartar resposta
		resp.Body.Close()
	}
}

func main() {
	start := time.Now()

	runtime.GOMAXPROCS(runtime.NumCPU()) // Utilizar todos os núcleos da máquina

	jobs := make(chan int, concurrentLimit)
	var wg sync.WaitGroup

	// Iniciar workers
	for i := 0; i < concurrentLimit; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	// Enviar jobs
	for i := 0; i < totalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	// Esperar finalização de todos os workers
	wg.Wait()

	fmt.Println("Completed in:", time.Since(start), "totalRequests:", totalRequests)
}
