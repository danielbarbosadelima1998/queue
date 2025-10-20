package main

import (
	"fmt"
	"log"
	"net"
	"queue/server"
	"queue/utils"
	"runtime"
	"time"
)

const (
	url = "localhost:7500"
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
	connPool, err := utils.NewConnPool(url, 100)

	if err != nil {
		log.Fatalf("failed on new connection pool, error: %v", err)
	}
	conns := make(map[int]net.Conn, 100)

	for i := 0; i < 100; i++ {
		conns[i] = connPool.Get()
	}

	return &LoadEnqueue{
		conns:      conns,
		connPool:   connPool,
		workerPool: utils.NewWorker("tcp-client", 100),
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	loadEnqueueu := NewLoadEnqueue()

	now := time.Now()

	loadEnqueueu.workerPool.Start()

	msg := []byte("Ping")

	for i := 0; i < 1000000; i++ {
		_, err := loadEnqueueu.workerPool.RunJob(loadEnqueueu.Job(msg))

		if err != nil {
			fmt.Println("failed to exec job", err.Error())
		}
	}

	loadEnqueueu.workerPool.CloseWorker()
	loadEnqueueu.connPool.Close()

	fmt.Println("time: ", time.Since(now))

}
