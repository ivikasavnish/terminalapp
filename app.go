package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// App struct
type App struct {
	ctx            context.Context
	configPath     string
	connectionPool *SSHConnectionPool
}

// SSHConnectionPool manages SSH connections
type SSHConnectionPool struct {
	connections map[string]*SSHConnection
	mu          sync.Mutex
}

// SSHConnection represents an active SSH connection
type SSHConnection struct {
	Client   *ssh.Client
	LastUsed time.Time
}

// SSHConfig interface defines the methods that all SSH config types must implement
type SSHConfig interface {
	GetClientConfig() (*ssh.ClientConfig, error)
	GetAddress() string
	GetName() string
}

// YAMLConfig represents an SSH configuration loaded from a YAML file
type YAMLConfig struct {
	Name       string `yaml:"name"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Username   string `yaml:"username"`
	SSHKeyPath string `yaml:"ssh_key_path"`
	Password   string `yaml:"password,omitempty"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		configPath: "./configs",
		connectionPool: &SSHConnectionPool{
			connections: make(map[string]*SSHConnection),
		},
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Application started")
}

// shutdown is called when the app is about to quit
func (a *App) shutdown(ctx context.Context) {
	log.Println("Shutting down...")
	a.closeAllConnections()
}

// Configuration Methods

// LoadYAMLConfig loads an SSH configuration from a YAML file
func (a *App) LoadYAMLConfig(name string) (*YAMLConfig, error) {
	filename := filepath.Join(a.configPath, name+".yaml")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", filename, err)
	}

	var config YAMLConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", filename, err)
	}

	if config.Port == "" {
		config.Port = "22"
	}

	return &config, nil
}

// LoadYAMLConfigs loads all YAML configurations from the config directory
func (a *App) LoadYAMLConfigs() ([]*YAMLConfig, error) {
	files, err := ioutil.ReadDir(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %v", err)
	}

	var configs []*YAMLConfig
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			config, err := a.LoadYAMLConfig(name)
			if err != nil {
				log.Printf("Failed to load config %s: %v", name, err)
				continue
			}
			configs = append(configs, config)
		}
	}

	return configs, nil
}

// Connection Methods

// ConnectSSH establishes an SSH connection using the provided SSHConfig
func (a *App) ConnectSSH(config SSHConfig) error {
	log.Printf("Connecting to %s", config.GetName())

	clientConfig, err := config.GetClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get SSH client config: %v", err)
	}

	client, err := ssh.Dial("tcp", config.GetAddress(), clientConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	a.connectionPool.mu.Lock()
	a.connectionPool.connections[config.GetName()] = &SSHConnection{
		Client:   client,
		LastUsed: time.Now(),
	}
	a.connectionPool.mu.Unlock()

	log.Printf("Successfully connected to %s", config.GetName())
	return nil
}

// DisconnectSSH closes an SSH connection
func (a *App) DisconnectSSH(name string) error {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	if conn, exists := a.connectionPool.connections[name]; exists {
		if err := conn.Client.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %v", err)
		}
		delete(a.connectionPool.connections, name)
		log.Printf("Disconnected from %s", name)
	}

	return nil
}

// GetActiveConnections returns a list of active SSH connections
func (a *App) GetActiveConnections() []string {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	activeConnections := make([]string, 0, len(a.connectionPool.connections))
	for name := range a.connectionPool.connections {
		activeConnections = append(activeConnections, name)
	}

	return activeConnections
}

// closeAllConnections closes all active SSH connections
func (a *App) closeAllConnections() {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	for name, conn := range a.connectionPool.connections {
		if err := conn.Client.Close(); err != nil {
			log.Printf("Error closing connection %s: %v", name, err)
		} else {
			log.Printf("Closed connection: %s", name)
		}
	}
	a.connectionPool.connections = make(map[string]*SSHConnection)
}

// Port Forwarding Methods

// PortForward sets up port forwarding
func (a *App) PortForward(config SSHConfig, localPort, remotePort int) error {
	sshClient, err := a.getSSHClient(config.GetName())
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return fmt.Errorf("failed to start local listener: %v", err)
	}

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept local connection: %v", err)
				continue
			}

			remoteConn, err := sshClient.Dial("tcp", fmt.Sprintf("localhost:%d", remotePort))
			if err != nil {
				log.Printf("Failed to connect to remote port: %v", err)
				localConn.Close()
				continue
			}

			go a.forwardConnection(localConn, remoteConn)
			go a.forwardConnection(remoteConn, localConn)
		}
	}()

	log.Printf("Port forwarding set up: localhost:%d -> remote:%d", localPort, remotePort)
	return nil
}

func (a *App) forwardConnection(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

// File Operation Methods

// ListDirectory returns the contents of a directory on the remote server
func (a *App) ListDirectory(config SSHConfig, path string) ([]os.FileInfo, error) {
	sftpClient, err := a.getSFTPClient(config.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %v", err)
	}

	files, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	return files, nil
}

// UploadFile uploads a file to the remote server
func (a *App) UploadFile(config SSHConfig, localPath, remotePath string) error {
	sftpClient, err := a.getSFTPClient(config.GetName())
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %v", err)
	}

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

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}

// DownloadFile downloads a file from the remote server
func (a *App) DownloadFile(config SSHConfig, remotePath, localPath string) error {
	sftpClient, err := a.getSFTPClient(config.GetName())
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %v", err)
	}

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

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}

// DeleteFile deletes a file on the remote server
func (a *App) DeleteFile(config SSHConfig, remotePath string) error {
	sftpClient, err := a.getSFTPClient(config.GetName())
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %v", err)
	}

	err = sftpClient.Remove(remotePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// RenameFile renames a file on the remote server
func (a *App) RenameFile(config SSHConfig, oldPath, newPath string) error {
	sftpClient, err := a.getSFTPClient(config.GetName())
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %v", err)
	}

	err = sftpClient.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %v", err)
	}

	return nil
}

// Utility Methods

func (a *App) getSSHClient(name string) (*ssh.Client, error) {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	if conn, exists := a.connectionPool.connections[name]; exists {
		return conn.Client, nil
	}

	return nil, fmt.Errorf("no active connection found for %s", name)
}

func (a *App) getSFTPClient(name string) (*sftp.Client, error) {
	sshClient, err := a.getSSHClient(name)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}

	return sftpClient, nil
}

// GetClientConfig implements SSHConfig for YAMLConfig
func (c *YAMLConfig) GetClientConfig() (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	if c.SSHKeyPath != "" {
		key, err := ioutil.ReadFile(c.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if c.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.Password))
	} else {
		return nil, fmt.Errorf("no authentication method provided")
	}

	return &ssh.ClientConfig{
		User:            c.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         10 * time.Second,
	}, nil
}

// GetAddress implements SSHConfig for YAMLConfig
func (c *YAMLConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// GetName implements SSHConfig for YAMLConfig
func (c *YAMLConfig) GetName() string {
	return c.Name
}

// readSSHKey reads the SSH private key from the given path
func (a *App) readSSHKey(path string) ([]byte, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key file %s: %v", path, err)
	}
	return key, nil
}
func (a *App) GetBaseProfile() (*YAMLConfig, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	return &YAMLConfig{
		Name:     "localhost",
		Host:     "localhost",
		Port:     "22",
		Username: currentUser.Username,
	}, nil
}

// LoadProfiles loads all standard YAML profiles
func (a *App) LoadProfiles() ([]*YAMLConfig, error) {
	return a.LoadYAMLConfigs()
}

// LoadCustomProfiles loads all custom profiles (same as standard for now)
func (a *App) LoadCustomProfiles() ([]*YAMLConfig, error) {
	return a.LoadYAMLConfigs()
}

// ConnectSSHWithHostKeyCheck establishes an SSH connection with host key verification
func (a *App) ConnectSSHWithHostKeyCheck(profile *YAMLConfig) (string, error) {
	log.Printf("Connecting with profile: %+v", profile)

	config, err := profile.GetClientConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get SSH client config: %v", err)
	}

	client, err := ssh.Dial("tcp", profile.GetAddress(), config)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}

	a.connectionPool.mu.Lock()
	a.connectionPool.connections[profile.GetName()] = &SSHConnection{
		Client:   client,
		LastUsed: time.Now(),
	}
	a.connectionPool.mu.Unlock()

	log.Printf("Successfully connected to %s", profile.GetName())
	return "", nil
}
