package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SavedCommand represents a single saved command
type SavedCommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// SavedCommandsManager manages the saved commands
type SavedCommandsManager struct {
	configPath string
	mu         sync.Mutex
}

// NewSavedCommandsManager creates a new SavedCommandsManager
func NewSavedCommandsManager(configPath string) *SavedCommandsManager {
	return &SavedCommandsManager{
		configPath: configPath,
	}
}

// ListSavedCommands retrieves all saved commands
func (scm *SavedCommandsManager) ListSavedCommands() ([]SavedCommand, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	savedCommandsPath := filepath.Join(scm.configPath, "saved_commands.json")

	if _, err := os.Stat(savedCommandsPath); os.IsNotExist(err) {
		return []SavedCommand{}, nil
	}

	data, err := ioutil.ReadFile(savedCommandsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read saved commands file: %v", err)
	}

	var commands []SavedCommand
	err = json.Unmarshal(data, &commands)
	if err != nil {
		return nil, fmt.Errorf("failed to parse saved commands: %v", err)
	}

	return commands, nil
}

// SaveCommand saves a new command
func (scm *SavedCommandsManager) SaveCommand(name string, command string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	commands, err := scm.ListSavedCommands()
	if err != nil {
		return err
	}

	// Check if command with the same name already exists
	for i, cmd := range commands {
		if cmd.Name == name {
			// Update existing command
			commands[i].Command = command
			return scm.saveCommandsToFile(commands)
		}
	}

	// Add new command
	commands = append(commands, SavedCommand{Name: name, Command: command})
	return scm.saveCommandsToFile(commands)
}

// DeleteSavedCommand deletes a saved command
func (scm *SavedCommandsManager) DeleteSavedCommand(name string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	commands, err := scm.ListSavedCommands()
	if err != nil {
		return err
	}

	var newCommands []SavedCommand
	for _, cmd := range commands {
		if cmd.Name != name {
			newCommands = append(newCommands, cmd)
		}
	}

	return scm.saveCommandsToFile(newCommands)
}

// saveCommandsToFile saves the commands to the file
func (scm *SavedCommandsManager) saveCommandsToFile(commands []SavedCommand) error {
	data, err := json.Marshal(commands)
	if err != nil {
		return fmt.Errorf("failed to marshal saved commands: %v", err)
	}

	savedCommandsPath := filepath.Join(scm.configPath, "saved_commands.json")
	err = ioutil.WriteFile(savedCommandsPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write saved commands file: %v", err)
	}

	return nil
}

// ExecuteSavedCommand executes a saved command on the specified profile
func (a *App) ExecuteSavedCommand(profile string, commandName string) error {
	commands, err := a.savedCommandsManager.ListSavedCommands()
	if err != nil {
		return fmt.Errorf("failed to list saved commands: %v", err)
	}

	var commandToExecute string
	for _, cmd := range commands {
		if cmd.Name == commandName {
			commandToExecute = cmd.Command
			break
		}
	}

	if commandToExecute == "" {
		return fmt.Errorf("command not found: %s", commandName)
	}

	return a.ExecuteInteractiveCommand(profile, commandToExecute)
}

func (a *App) readAndEmitOutput(profile string, reader io.Reader, outputType string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		runtime.EventsEmit(a.ctx, "command_output", map[string]string{
			"profile": profile,
			"type":    outputType,
			"data":    scanner.Text(),
		})
	}
}

// These methods should be added to the App struct to interface with SavedCommandsManager

func (a *App) ListSavedCommands() ([]SavedCommand, error) {
	return a.savedCommandsManager.ListSavedCommands()
}

func (a *App) SaveCommand(name string, command string) error {
	return a.savedCommandsManager.SaveCommand(name, command)
}

func (a *App) DeleteSavedCommand(name string) error {
	return a.savedCommandsManager.DeleteSavedCommand(name)
}
