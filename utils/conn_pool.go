package utils

import (
	"fmt"
	"net"
	"sync"
)

type ConnPool struct {
	conns chan net.Conn
	mu    sync.Mutex
	addr  string
}

func NewConnPool(addr string, size int) (*ConnPool, error) {
	pool := &ConnPool{
		conns: make(chan net.Conn, size),
		addr:  addr,
	}

	for i := 0; i < size; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %v", err)
		}
		pool.conns <- conn
	}

	return pool, nil
}

// Pega uma conexão do pool (bloqueia se não tiver disponível)
func (p *ConnPool) Get() net.Conn {
	return <-p.conns
}

// Devolve uma conexão para o pool
func (p *ConnPool) Put(conn net.Conn) {
	p.conns <- conn
}

// Fecha todas as conexões do pool
func (p *ConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}
}
