package main

import (
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type PortForward struct {
	LocalPort       int  `json:"localPort"`
	RemotePort      int  `json:"remotePort"`
	IsRemoteToLocal bool `json:"isRemoteToLocal"`
}

var (
	activeForwards      = make(map[string][]*PortForward)
	activeForwardsMutex sync.Mutex
)

func (a *App) PortForward(profile string, localPort, remotePort int, isRemoteToLocal bool) error {
	a.connectionPool.mu.Lock()
	conn, exists := a.connectionPool.connections[profile]
	a.connectionPool.mu.Unlock()

	if !exists {
		return fmt.Errorf("no active connection found for profile: %s", profile)
	}

	var listener net.Listener
	var err error

	if isRemoteToLocal {
		// Remote to local port forwarding
		listener, err = conn.Client.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", remotePort))
	} else {
		// Local to remote port forwarding
		listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	}

	if err != nil {
		return fmt.Errorf("failed to set up port forwarding: %v", err)
	}

	forward := &PortForward{
		LocalPort:       localPort,
		RemotePort:      remotePort,
		IsRemoteToLocal: isRemoteToLocal,
	}

	activeForwardsMutex.Lock()
	activeForwards[profile] = append(activeForwards[profile], forward)
	activeForwardsMutex.Unlock()

	go a.handlePortForward(listener, conn.Client, localPort, remotePort, isRemoteToLocal)

	return nil
}

func (a *App) handlePortForward(listener net.Listener, sshClient *ssh.Client, localPort, remotePort int, isRemoteToLocal bool) {
	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			return
		}

		go func() {
			defer localConn.Close()

			var remoteConn net.Conn
			var err error

			if isRemoteToLocal {
				remoteConn, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
			} else {
				remoteConn, err = sshClient.Dial("tcp", fmt.Sprintf("localhost:%d", remotePort))
			}

			if err != nil {
				fmt.Printf("Failed to connect to remote: %v\n", err)
				return
			}
			defer remoteConn.Close()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				_, _ = copyIO(localConn, remoteConn)
			}()

			go func() {
				defer wg.Done()
				_, _ = copyIO(remoteConn, localConn)
			}()

			wg.Wait()
		}()
	}
}

func copyIO(dst, src net.Conn) (int64, error) {
	return dst.(*net.TCPConn).ReadFrom(src)
}

func (a *App) StopPortForward(profile string, localPort, remotePort int, isRemoteToLocal bool) error {
	activeForwardsMutex.Lock()
	defer activeForwardsMutex.Unlock()

	forwards, exists := activeForwards[profile]
	if !exists {
		return fmt.Errorf("no active port forwards found for profile: %s", profile)
	}

	for i, forward := range forwards {
		if forward.LocalPort == localPort && forward.RemotePort == remotePort && forward.IsRemoteToLocal == isRemoteToLocal {
			// Remove the forward from the slice
			activeForwards[profile] = append(forwards[:i], forwards[i+1:]...)
			// TODO: Implement actual stopping of the port forward
			return nil
		}
	}

	return fmt.Errorf("port forward not found")
}

func (a *App) GetActivePortForwards(profile string) ([]*PortForward, error) {
	activeForwardsMutex.Lock()
	defer activeForwardsMutex.Unlock()

	forwards, exists := activeForwards[profile]
	if !exists {
		return []*PortForward{}, nil
	}

	return forwards, nil
}
