package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type SSHConnection struct {
	Client   *ssh.Client
	LastUsed time.Time
}

type SSHConnectionPool struct {
	connections map[string]*SSHConnection
	mu          sync.Mutex
	maxIdleTime time.Duration
}

func NewSSHConnectionPool(maxIdleTime time.Duration) *SSHConnectionPool {
	pool := &SSHConnectionPool{
		connections: make(map[string]*SSHConnection),
		maxIdleTime: maxIdleTime,
	}

	go pool.cleanupIdleConnections()

	return pool
}

func (p *SSHConnectionPool) GetConnection(profile string, config *ssh.ClientConfig, address string) (*ssh.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := fmt.Sprintf("%s-%s", profile, address)

	if conn, exists := p.connections[key]; exists {
		conn.LastUsed = time.Now()
		return conn.Client, nil
	}

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", address, err)
	}

	p.connections[key] = &SSHConnection{
		Client:   client,
		LastUsed: time.Now(),
	}

	return client, nil
}

func (p *SSHConnectionPool) CloseConnection(profile string, address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := fmt.Sprintf("%s-%s", profile, address)

	if conn, exists := p.connections[key]; exists {
		if err := conn.Client.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %v", err)
		}
		delete(p.connections, key)
	}

	return nil
}

func (p *SSHConnectionPool) cleanupIdleConnections() {
	for {
		time.Sleep(p.maxIdleTime)

		p.mu.Lock()
		for key, conn := range p.connections {
			if time.Since(conn.LastUsed) > p.maxIdleTime {
				conn.Client.Close()
				delete(p.connections, key)
			}
		}
		p.mu.Unlock()
	}
}
