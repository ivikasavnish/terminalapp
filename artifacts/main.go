package main

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
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

// App struct
type App struct {
	ctx            context.Context
	configPath     string
	connectionPool *SSHConnectionPool
	savedCommands  map[string]string
	mu             sync.Mutex
}

// SSHConfig holds the configuration for an SSH connection
type SSHConfig struct {
	SSHKeyPath string `yaml:"ssh_key_path"`
	Username   string `yaml:"username"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		configPath:     "./configs",
		connectionPool: NewSSHConnectionPool(5 * time.Minute),
		savedCommands:  make(map[string]string),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SetConfigPath sets the path for the config directory
func (a *App) SetConfigPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config path does not exist: %s", path)
	}
	a.configPath = path
	return nil
}

// LoadProfiles loads all SSH profiles from the config directory
func (a *App) LoadProfiles() ([]string, error) {
	files, err := ioutil.ReadDir(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %v", err)
	}

	var profiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			profiles = append(profiles, file.Name()[:len(file.Name())-5])
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles found in config directory")
	}

	return profiles, nil
}

// loadSSHConfig loads the SSH configuration for a given profile
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

// ConnectSSH establishes an SSH connection for a given profile
func (a *App) ConnectSSH(profile string) error {
	config, err := a.loadSSHConfig(profile)
	if err != nil {
		return err
	}

	key, err := ioutil.ReadFile(config.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %v", err)
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
	_, err = a.connectionPool.GetConnection(profile, clientConfig, address)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	return nil
}

// DisconnectSSH closes an SSH connection for a given profile
func (a *App) DisconnectSSH(profile string) error {
	config, err := a.loadSSHConfig(profile)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	return a.connectionPool.CloseConnection(profile, address)
}

// GetActiveConnections returns a list of profiles with active connections
func (a *App) GetActiveConnections() []string {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	activeProfiles := make([]string, 0, len(a.connectionPool.connections))
	for key := range a.connectionPool.connections {
		profile := key[:strings.Index(key, "-")]
		activeProfiles = append(activeProfiles, profile)
	}
	return activeProfiles
}

// ExecuteCommand executes a command on the remote server
func (a *App) ExecuteCommand(profile, command string) (string, error) {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// forwardConnection handles the forwarding of data between two connections
func (a *App) forwardConnection(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

// SaveCommand saves a command with a given name
func (a *App) SaveCommand(name, command string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.savedCommands[name] = command
	return nil
}

// GetSavedCommand retrieves a saved command by name
func (a *App) GetSavedCommand(name string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	command, exists := a.savedCommands[name]
	if !exists {
		return "", fmt.Errorf("no saved command found with name: %s", name)
	}

	return command, nil
}

// ListSavedCommands returns all saved commands
func (a *App) ListSavedCommands() map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()

	commands := make(map[string]string)
	for name, command := range a.savedCommands {
		commands[name] = command
	}

	return commands
}

// ExecuteSavedCommand executes a saved command
func (a *App) ExecuteSavedCommand(profile, name string) (string, error) {
	command, err := a.GetSavedCommand(name)
	if err != nil {
		return "", err
	}

	return a.ExecuteCommand(profile, command)
}

// CreateRemoteDirectory creates a new directory on the remote server
func (a *App) CreateRemoteDirectory(profile, path string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}

	defer sftpClient.Close()

	return sftpClient.MkdirAll(path)
}

// DeleteRemoteFile deletes a file or empty directory on the remote server
func (a *App) DeleteRemoteFile(profile, path string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	return sftpClient.Remove(path)
}

// RenameRemoteFile renames a file or directory on the remote server
func (a *App) RenameRemoteFile(profile, oldPath, newPath string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	return sftpClient.Rename(oldPath, newPath)
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

type InteractiveCommand struct {
	Session *ssh.Session
	Stdout  io.Reader
	Stderr  io.Reader
	Cancel  context.CancelFunc
}

var (
	interactiveCommands = make(map[string]*InteractiveCommand)
	commandMutex        sync.Mutex
)

func (a *App) ExecuteInteractiveCommand(profile, command string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := session.Start(command); err != nil {
		session.Close()
		return fmt.Errorf("failed to start command: %v", err)
	}

	ctx, cancel := context.WithCancel(a.ctx)

	ic := &InteractiveCommand{
		Session: session,
		Stdout:  stdout,
		Stderr:  stderr,
		Cancel:  cancel,
	}

	commandMutex.Lock()
	interactiveCommands[profile] = ic
	commandMutex.Unlock()

	go func() {
		defer session.Close()
		defer func() {
			commandMutex.Lock()
			delete(interactiveCommands, profile)
			commandMutex.Unlock()
		}()

		stdoutDone := make(chan bool)
		stderrDone := make(chan bool)

		go a.streamOutput(ctx, stdout, "stdout", profile, stdoutDone)
		go a.streamOutput(ctx, stderr, "stderr", profile, stderrDone)

		select {
		case <-ctx.Done():
			return
		case <-stdoutDone:
			<-stderrDone
		case <-stderrDone:
			<-stdoutDone
		}

		if err := session.Wait(); err != nil {
			runtime.EventsEmit(a.ctx, "command_output", map[string]string{
				"profile": profile,
				"type":    "error",
				"data":    fmt.Sprintf("Command finished with error: %v", err),
			})
		} else {
			runtime.EventsEmit(a.ctx, "command_output", map[string]string{
				"profile": profile,
				"type":    "info",
				"data":    "Command finished successfully",
			})
		}
	}()

	return nil
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

func (a *App) StopInteractiveCommand(profile string) error {
	commandMutex.Lock()
	ic, exists := interactiveCommands[profile]
	commandMutex.Unlock()

	if !exists {
		return fmt.Errorf("no interactive command running for profile: %s", profile)
	}

	// Send SIGINT
	if err := ic.Session.Signal(ssh.SIGINT); err != nil {
		log.Printf("Failed to send SIGINT: %v", err)
	}

	// Wait for a short period to see if the process exits gracefully
	select {
	case <-time.After(2 * time.Second):
		// If it doesn't exit, force close the session
		ic.Session.Close()
	default:
		// Do nothing

	}

	ic.Cancel() // Cancel the context

	commandMutex.Lock()
	delete(interactiveCommands, profile)
	commandMutex.Unlock()

	runtime.EventsEmit(a.ctx, "command_output", map[string]string{
		"profile": profile,
		"type":    "info",
		"data":    "Command stopped",
	})

	return nil
}

func (a *App) PortForward(profile string, localPort, remotePort int, isRemoteToLocal bool) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	var listener net.Listener
	var remoteAddr string

	if isRemoteToLocal {
		// Remote to local port forwarding
		remoteAddr = fmt.Sprintf("0.0.0.0:%d", remotePort) // Listen on all interfaces
		listener, err = client.Listen("tcp", remoteAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on remote port: %v", err)
		}
	} else {
		// Local to remote port forwarding
		localAddr := fmt.Sprintf("localhost:%d", localPort)
		listener, err = net.Listen("tcp", localAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on local port: %v", err)
		}
	}

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				return
			}

			go func() {
				var remote net.Conn
				var err error

				if isRemoteToLocal {
					remote, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
				} else {
					remote, err = client.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", remotePort)) // Allow external access
				}

				if err != nil {
					log.Printf("Failed to connect to %s: %v", remoteAddr, err)
					conn.Close()
					return
				}

				go a.forwardConnection(remote, conn)
				go a.forwardConnection(conn, remote)
			}()
		}
	}()

	directionStr := "local to remote"
	if isRemoteToLocal {
		directionStr = "remote to local"
	}
	log.Printf("Port forwarding (%s) set up: %d <-> %d", directionStr, localPort, remotePort)

	return nil
}

// ProgressReader is a wrapper for io.Reader that reports progress
type ProgressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress func(float64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.OnProgress != nil {
		pr.OnProgress(float64(pr.Current) / float64(pr.Total) * 100)
	}
	return n, err
}

// ListDirectory returns detailed file information for a directory
func (a *App) ListDirectory(profile, path string) ([]map[string]interface{}, error) {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var fileList []map[string]interface{}
	for _, file := range files {
		fileList = append(fileList, map[string]interface{}{
			"name":  file.Name(),
			"isDir": file.IsDir(),
			"size":  file.Size(),
			"mode":  file.Mode().String(),
		})
	}

	return fileList, nil
}
func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select file to upload",
	})
}

func (a *App) SaveFileDialog(defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save file",
		DefaultFilename: defaultFilename,
	})
}

func (a *App) UploadFile(profile, localPath, remotePath string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	total := 0
	for {
		n, err := localFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := remoteFile.Write(buf[:n]); err != nil {
			return err
		}

		total += n
		runtime.EventsEmit(a.ctx, "file_progress", map[string]interface{}{
			"operation": "upload",
			"filename":  filepath.Base(localPath),
			"progress":  float64(total) / float64(1024*1024), // Assuming 1MB total size for simplicity
		})
	}

	return nil
}

func (a *App) DownloadFile(profile, remotePath, localPath string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	total := 0
	for {
		n, err := remoteFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := localFile.Write(buf[:n]); err != nil {
			return err
		}

		total += n
		runtime.EventsEmit(a.ctx, "file_progress", map[string]interface{}{
			"operation": "download",
			"filename":  filepath.Base(remotePath),
			"progress":  float64(total) / float64(1024*1024), // Assuming 1MB total size for simplicity
		})
	}

	return nil
}
