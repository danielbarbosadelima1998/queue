package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"queue/utils"
)

type TcpServer struct {
	addr           string
	ProcessMessage func([]byte)
	workerPool     *utils.Worker
}

func NewTcpServer(processor func([]byte)) *TcpServer {
	port := os.Getenv("API_PORT")

	if port == "" {
		log.Fatal("failed to start http server, port is empty")
	}

	address := ":" + port
	return &TcpServer{
		addr:           address,
		ProcessMessage: processor,
		workerPool:     utils.NewWorker("tcp-server", 100),
	}
}

func (s *TcpServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("erro ao iniciar listener: %w", err)
	}
	defer ln.Close()

	fmt.Printf("Tcp server listen on %s\n", s.addr)

	s.workerPool.Start()
	defer s.workerPool.CloseWorker()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao aceitar conexão: %v\n", err)
			continue
		}

		_, _ = s.workerPool.RunJob(func(workerNumber int) (response any, err error) {
			s.handleConnection(conn)

			return nil, nil
		})
	}
}

func (s *TcpServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	sizeBuf := make([]byte, 4)

	for {
		// 1. LÊ O TAMANHO DA MENSAGEM
		_, err := io.ReadFull(reader, sizeBuf)
		if err != nil {
			return
		}

		msgSize := binary.LittleEndian.Uint32(sizeBuf)

		// 2. LÊ O CORPO DA MENSAGEM
		payload := make([]byte, msgSize)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return
		}

		// 3. PROCESSA MENSAGEM
		s.ProcessMessage(payload)
	}
}

func SendRequest(conn net.Conn, payload []byte) (response any, err error) {
	msgSize := uint32(len(payload))

	// 1. ESCREVE O TAMANHO DA MENSAGEM
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, msgSize)

	if _, err := conn.Write(sizeBuf); err != nil {
		return nil, fmt.Errorf("erro ao escrever o tamanho: %w", err)
	}

	// 2. ESCREVE O CORPO DA MENSAGEM
	if _, err := conn.Write(payload); err != nil {
		return nil, fmt.Errorf("erro ao escrever o payload: %w", err)
	}

	return nil, nil
}
