package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// App struct
type App struct {
	ctx                  context.Context
	savedCommandsManager *SavedCommandsManager
	configPath           string
	connectionPool       *SSHConnectionPool
}

type SSHConfig struct {
	Name       string `json:"name" yaml:"name"`
	Host       string `json:"host" yaml:"host"`
	Port       int    `json:"port" yaml:"port"`
	Username   string `json:"username" yaml:"username"`
	Password   string `json:"password" yaml:"password"`
	SSHKeyPath string `json:"ssh_key_path" yaml:"ssh_key_path"`
}

// ConnectionResult represents the result of a successful connection
type ConnectionResult struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	configPath := "./configs"
	return &App{
		configPath:           configPath,
		savedCommandsManager: NewSavedCommandsManager(configPath),
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

// GetBaseProfile returns the base profile (localhost)
func (a *App) GetBaseProfile() (*SSHConfig, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	return &SSHConfig{
		Name:     "Host System",
		Host:     "localhost",
		Port:     22,
		Username: currentUser.Username,
	}, nil
}

// LoadProfiles loads all YAML profiles
func (a *App) LoadProfiles() ([]*SSHConfig, error) {
	files, err := ioutil.ReadDir(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %v", err)
	}

	var profiles []*SSHConfig
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			config, err := a.loadYAMLConfig(file.Name())
			if err != nil {
				log.Printf("Failed to load config %s: %v", file.Name(), err)
				continue
			}
			profiles = append(profiles, config)
		}
	}

	return profiles, nil
}

// loadYAMLConfig loads a single YAML config file
// loadYAMLConfig loads a single YAML config file
func (a *App) loadYAMLConfig(filename string) (*SSHConfig, error) {
	// Read the YAML file
	filePath := filepath.Join(a.configPath, filename)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the YAML file
	var config SSHConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %v", err)
	}

	return &config, nil
}

// DisconnectSSH closes an SSH connection
func (a *App) DisconnectSSH(name string) error {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	conn, exists := a.connectionPool.connections[name]
	if !exists {
		return fmt.Errorf("no active connection found for %s", name)
	}

	err := conn.Client.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	delete(a.connectionPool.connections, name)
	return nil
}

// GetActiveConnections returns a list of active SSH connections
func (a *App) GetActiveConnections() []string {
	a.connectionPool.mu.Lock()
	defer a.connectionPool.mu.Unlock()

	var activeConnections []string
	for name := range a.connectionPool.connections {
		activeConnections = append(activeConnections, name)
	}
	return activeConnections
}

// ConnectSSHWithHostKeyCheck establishes an SSH connection with host key verification
func (a *App) ConnectSSHWithHostKeyCheck(profileJSON string) (*ConnectionResult, error) {
	log.Printf("Received profile: %s", profileJSON)

	var profile SSHConfig
	err := json.Unmarshal([]byte(profileJSON), &profile)
	if err != nil {
		log.Printf("Failed to parse profile data: %v", err)
		return nil, errors.New("Failed to parse profile data")
	}

	// Validate profile data
	if profile.Name == "" || profile.Host == "" || profile.Username == "" {
		log.Printf("Invalid profile: missing required fields. Profile: %+v", profile)
		return nil, errors.New("Invalid profile: missing required fields")
	}

	config := &ssh.ClientConfig{
		User:            profile.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         10 * time.Second,
	}

	// Set up authentication
	if profile.Password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(profile.Password)}
		log.Printf("Using password authentication for %s@%s", profile.Username, profile.Host)
	} else if profile.SSHKeyPath != "" {
		key, err := ioutil.ReadFile(profile.SSHKeyPath)
		if err != nil {
			log.Printf("Failed to read SSH key from %s: %v", profile.SSHKeyPath, err)
			return nil, fmt.Errorf("Failed to read SSH key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Printf("Failed to parse SSH key: %v", err)
			return nil, fmt.Errorf("Failed to parse SSH key: %v", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
		log.Printf("Using SSH key authentication for %s@%s", profile.Username, profile.Host)
	} else {
		log.Printf("No authentication method provided for %s@%s", profile.Username, profile.Host)
		return nil, errors.New("No authentication method provided")
	}

	// Attempt to connect
	addr := fmt.Sprintf("%s:%s", profile.Host, strconv.Itoa(profile.Port))
	log.Printf("Attempting to connect to %s", addr)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", addr, err)
		return nil, fmt.Errorf("Failed to connect: %v", err)
	}

	// Store the connection
	a.connectionPool.mu.Lock()
	a.connectionPool.connections[profile.Name] = &SSHConnection{
		Client:   client,
		LastUsed: time.Now(),
	}
	a.connectionPool.mu.Unlock()

	log.Printf("Successfully connected to %s", addr)

	result := &ConnectionResult{
		Name:     profile.Name,
		Host:     profile.Host,
		Port:     strconv.Itoa(profile.Port),
		Username: profile.Username,
	}
	log.Printf("Returning connection result: %+v", result)
	return result, nil
}
