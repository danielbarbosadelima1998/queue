package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"queue/server"
	"queue/utils"
	"runtime"
	"time"

	"github.com/joho/godotenv"
)

const (
	requestCount = 1_000_000
	concurrency  = 100
)

func (l *LoadEnqueue) Job(msg []byte) utils.Job {
	return func(workerNumber int) (response any, err error) {
		response, err = server.SendRequest(l.conns[workerNumber], msg)

		if err != nil {
			fmt.Println("err", err)
		}
		return response, nil
	}
}

type LoadEnqueue struct {
	conns      map[int]net.Conn
	connPool   *utils.ConnPool
	workerPool *utils.Worker
}

func NewLoadEnqueue() *LoadEnqueue {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file ", err.Error())
	}

	url := os.Getenv("QUEUE_URL")

	if url == "" {
		log.Fatal("QUEUE_URL is empty")
	}

	connPool, err := utils.NewConnPool(url, concurrency)

	if err != nil {
		log.Fatalf("failed on new connection pool, error: %v", err)
	}
	conns := make(map[int]net.Conn, concurrency)

	for i := 0; i < concurrency; i++ {
		conns[i] = connPool.Get()
	}

	return &LoadEnqueue{
		conns:      conns,
		connPool:   connPool,
		workerPool: utils.NewWorker("tcp-client", concurrency),
	}
}

func main() {
	fmt.Printf("Running load test for tcp requests with concurrency %d\n for worker pool and tcp connection pool, request count: %d\n", concurrency, requestCount)

	runtime.GOMAXPROCS(runtime.NumCPU())

	loadEnqueueu := NewLoadEnqueue()

	now := time.Now()

	loadEnqueueu.workerPool.Start()

	msg := []byte("Ping")

	for i := 0; i < requestCount; i++ {
		_, err := loadEnqueueu.workerPool.RunJob(loadEnqueueu.Job(msg))

		if err != nil {
			fmt.Println("failed to exec job", err.Error())
		}
	}

	loadEnqueueu.workerPool.CloseWorker()
	loadEnqueueu.connPool.Close()

	fmt.Println("Request count:", requestCount)

	elapsed := time.Since(now)
	fmt.Println("Total time:", elapsed)

	reqPerSec := float64(requestCount) / elapsed.Seconds()
	fmt.Printf("Req/seg: %.2f\n", reqPerSec)

	latencyMs := elapsed.Seconds() * 1000 / float64(requestCount)
	fmt.Printf("Latência média (ms): %.3f\n", latencyMs)

	// System Information
	cmd := exec.Command("sh", "-c", "grep 'model name' /proc/cpuinfo | head -1")
	cpuInfo, _ := cmd.Output()

	cmd = exec.Command("sh", "-c", "grep 'MemTotal' /proc/meminfo")
	memoryInfo, _ := cmd.Output()

	fmt.Println("cpuInfo: ", string(cpuInfo))
	fmt.Println("memoryInfo: ", string(memoryInfo))
}
