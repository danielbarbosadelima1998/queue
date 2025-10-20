package main

import (
	"fmt"
	"log"
	"queue/server"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file ", err.Error())
	}

	// store := store.NewMemoryStore()

	// httpServer := server.NewHttpServer(store)

	// log.Fatal(httpServer.Start())

	tcpServer := server.NewTcpServer(func(b []byte) {
		// @TODO: Implementar "rotas" enqueue, dequeue, etc.
		// fmt.Println("new message received!", string(b))
	})

	log.Fatal(tcpServer.Start())
}
