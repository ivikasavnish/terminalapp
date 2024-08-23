package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	maxHistorySize = 100
	historyDir     = "./history"
)

var (
	historyMutex sync.Mutex
	synonyms     = make(map[string]string)
)

func (a *App) GetCommandHistory(profile string) ([]string, error) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	filename := filepath.Join(historyDir, fmt.Sprintf("%s_history.txt", profile))
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	var history []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read history file: %v", err)
	}

	// Reverse the history so that the most recent commands are first
	for i := len(history)/2 - 1; i >= 0; i-- {
		opp := len(history) - 1 - i
		history[i], history[opp] = history[opp], history[i]
	}

	return history, nil
}

func (a *App) AddCommandToHistory(profile string, command string) error {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %v", err)
	}

	filename := filepath.Join(historyDir, fmt.Sprintf("%s_history.txt", profile))
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(command + "\n"); err != nil {
		return fmt.Errorf("failed to write to history file: %v", err)
	}

	// Trim history if it exceeds maxHistorySize
	if err := a.trimHistory(filename); err != nil {
		return fmt.Errorf("failed to trim history: %v", err)
	}

	return nil
}

func (a *App) trimHistory(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > maxHistorySize {
		lines = lines[len(lines)-maxHistorySize:]
		return os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
	}

	return nil
}

func (a *App) CreateSynonym(command string) (string, error) {
	words := strings.Fields(command)
	if len(words) < 2 {
		return "", nil // No synonym for short commands
	}

	// Create a simple acronym-based synonym
	acronym := ""
	for _, word := range words {
		if len(word) > 0 {
			acronym += string(word[0])
		}
	}

	// If the acronym is already in use, append a number
	baseSynonym := acronym
	count := 1
	for {
		if _, exists := synonyms[acronym]; !exists {
			break
		}
		acronym = fmt.Sprintf("%s%d", baseSynonym, count)
		count++
	}

	synonyms[acronym] = command

	// Save synonyms to a file
	if err := a.saveSynonyms(); err != nil {
		return "", fmt.Errorf("failed to save synonym: %v", err)
	}

	return acronym, nil
}

func (a *App) saveSynonyms() error {
	file, err := os.Create(filepath.Join(historyDir, "synonyms.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(synonyms)
}

func (a *App) loadSynonyms() error {
	file, err := os.Open(filepath.Join(historyDir, "synonyms.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil // It's okay if the file doesn't exist yet
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&synonyms)
}
