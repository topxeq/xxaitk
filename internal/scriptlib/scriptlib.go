package scriptlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var RegistryURL = "https://raw.githubusercontent.com/topxeq/xxaitk-scripts/main/registry.json"

type Library struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
}

type Registry struct {
	Libraries []Library `json:"libraries"`
}

func LibDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aitk", "libs")
}

func ListInstalled() ([]string, error) {
	dir := LibDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".xxs") {
			names = append(names, strings.TrimSuffix(e.Name(), ".xxs"))
		}
	}
	return names, nil
}

func ListRemote() (*Registry, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(RegistryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(body, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}
	return &reg, nil
}

func Get(name string) error {
	reg, err := ListRemote()
	if err != nil {
		return err
	}

	var lib *Library
	for i := range reg.Libraries {
		if reg.Libraries[i].Name == name {
			lib = &reg.Libraries[i]
			break
		}
	}
	if lib == nil {
		return fmt.Errorf("library not found: %s", name)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(lib.URL)
	if err != nil {
		return fmt.Errorf("failed to download library: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read library: %w", err)
	}

	dir := LibDir()
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, name+".xxs")
	if err := os.WriteFile(path, body, 0644); err != nil {
		return fmt.Errorf("failed to write library: %w", err)
	}

	return nil
}

func Remove(name string) error {
	path := filepath.Join(LibDir(), name+".xxs")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("library not installed: %s", name)
	}
	return os.Remove(path)
}

func Load(name string) (string, error) {
	path := filepath.Join(LibDir(), name+".xxs")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load library %s: %w", name, err)
	}
	return string(data), nil
}
