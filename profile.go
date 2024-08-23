package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Profile struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

type CustomProfile struct {
	Profile
	Password string `json:"password"`
}

func (a *App) loadProfile(filename string) (Profile, error) {
	data, err := ioutil.ReadFile(filepath.Join(a.configPath, filename))
	if err != nil {
		return Profile{}, fmt.Errorf("failed to read profile file %s: %v", filename, err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return Profile{}, fmt.Errorf("failed to parse profile file %s: %v", filename, err)
	}

	profile.Name = filename[:len(filename)-5] // Remove .yaml extension
	return profile, nil
}

func (a *App) SaveCustomProfile(profile CustomProfile) error {
	customProfilesDir := filepath.Join(a.configPath, "custom_profiles")
	if err := os.MkdirAll(customProfilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom profiles directory: %v", err)
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal custom profile: %v", err)
	}

	filename := filepath.Join(customProfilesDir, profile.Name+".json")
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write custom profile file: %v", err)
	}

	return nil
}

func (a *App) DeleteCustomProfile(profileName string) error {
	filename := filepath.Join(a.configPath, "custom_profiles", profileName+".json")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete custom profile %s: %v", profileName, err)
	}
	return nil
}
