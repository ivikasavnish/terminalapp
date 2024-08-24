package main

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	activeSessionsMutex sync.Mutex
	activeSessions      = make(map[string]*ssh.Session)
)

type InteractiveCommand struct {
	Session *ssh.Session
	Stdout  io.Reader
	Stderr  io.Reader
	Cancel  context.CancelFunc
}

func (a *App) loadSSHConfig(profile string) (*SSHConfig, error) {
	filename := filepath.Join(a.configPath, profile+".yaml")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", filename, err)
	}

	var config SSHConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", filename, err)
	}

	// Expand ~ to home directory if present
	if config.SSHKeyPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		config.SSHKeyPath = filepath.Join(home, config.SSHKeyPath[1:])
	}

	return &config, nil
}

// getSSHClient is a helper function to get the SSH client for a profile
func (a *App) getSSHClient(profile string) (*ssh.Client, error) {
	config, err := a.loadSSHConfig(profile)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(config.SSHKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	clientConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	return a.connectionPool.GetConnection(profile, clientConfig, address)
}

var (
	interactiveCommands = make(map[string]*InteractiveCommand)
	commandMutex        sync.Mutex
)

func (a *App) emitOutput(profile, outputType, data string) {
	if a.ctx == nil {
		fmt.Printf("Warning: Context is nil, unable to emit event. Profile: %s, Type: %s, Data: %s\n", profile, outputType, data)
		return
	}

	event := map[string]interface{}{
		"profile": profile,
		"type":    outputType,
		"data":    data,
	}

	// Emit the event to the frontend
	runtime.EventsEmit(a.ctx, "command_output", event)

	// Also print to console for debugging
	fmt.Printf("Emitted - Profile: %s, Type: %s, Data: %s\n", profile, outputType, data)
}
func (a *App) streamOutput(ctx context.Context, r io.Reader, outputType string, profile string, done chan<- bool) {
	defer close(done)
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := r.Read(buf)
			if n > 0 {
				runtime.EventsEmit(a.ctx, "command_output", map[string]string{
					"profile": profile,
					"type":    outputType,
					"data":    string(buf[:n]),
				})
			}
			if err != nil {
				if err != io.EOF {
					runtime.EventsEmit(a.ctx, "command_output", map[string]string{
						"profile": profile,
						"type":    "error",
						"data":    fmt.Sprintf("Error reading output: %v", err),
					})
				}
				return
			}
		}
	}
}

func (a *App) ExecuteInteractiveCommand(profile, command string) error {
	// Check for "clear" command
	if strings.TrimSpace(strings.ToLower(command)) == "clear" {
		a.emitOutput(profile, "clear", "")
		return nil
	}

	// Rest of the function remains the same
	a.connectionPool.mu.Lock()
	conn, exists := a.connectionPool.connections[profile]
	a.connectionPool.mu.Unlock()

	if !exists {
		return fmt.Errorf("no active connection found for profile: %s", profile)
	}

	session, err := conn.Client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	activeSessionsMutex.Lock()
	activeSessions[profile] = session
	activeSessionsMutex.Unlock()

	defer func() {
		session.Close()
		activeSessionsMutex.Lock()
		delete(activeSessions, profile)
		activeSessionsMutex.Unlock()
	}()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %v", err)
	}

	outputDone := make(chan bool)
	go a.handleOutput(profile, stdout, "stdout", outputDone)
	go a.handleOutput(profile, stderr, "stderr", outputDone)

	_, err = fmt.Fprintln(stdin, command)
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	stdin.Close()

	err = session.Wait()

	<-outputDone
	<-outputDone

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			a.emitOutput(profile, "error", fmt.Sprintf("Command exited with code %d", exitErr.ExitStatus()))
		} else {
			a.emitOutput(profile, "error", fmt.Sprintf("Command failed: %v", err))
		}
	} else {
		a.emitOutput(profile, "info", "Command finished successfully")
	}

	return nil
}

func (a *App) handleOutput(profile string, r io.Reader, outputType string, done chan<- bool) {
	defer func() { done <- true }()

	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			a.emitOutput(profile, outputType, string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				a.emitOutput(profile, "error", fmt.Sprintf("Error reading %s: %v", outputType, err))
			}
			break
		}
	}
}

func (a *App) StopInteractiveCommand(profile string) error {
	activeSessionsMutex.Lock()
	session, exists := activeSessions[profile]
	activeSessionsMutex.Unlock()

	if !exists {
		return fmt.Errorf("no active session found for profile: %s", profile)
	}

	// Send SIGINT to the remote process
	if err := session.Signal(ssh.SIGINT); err != nil {
		return fmt.Errorf("failed to send SIGINT: %v", err)
	}

	// Give the command a moment to stop gracefully
	time.Sleep(time.Second)

	// If it's still running, force close the session
	activeSessionsMutex.Lock()
	if _, stillExists := activeSessions[profile]; stillExists {
		session.Close()
		delete(activeSessions, profile)
	}
	activeSessionsMutex.Unlock()

	a.emitOutput(profile, "info", "Command stopped")
	return nil
}
