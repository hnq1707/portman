package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nay-kia/portman/internal/port"
)

// Profile represents a saved port profile.
type Profile struct {
	Name      string      `json:"name"`
	Ports     []PortEntry `json:"ports"`
	CreatedAt time.Time   `json:"created_at"`
}

// PortEntry is a single port in a profile.
type PortEntry struct {
	Port        int    `json:"port"`
	Proto       string `json:"proto"`
	ProcessName string `json:"process_name"`
}

// CheckResult holds the result of checking a port.
type CheckResult struct {
	Entry   PortEntry
	Active  bool
	Current string // current process name if active
}

func profileDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".portman", "profiles")
	return dir, os.MkdirAll(dir, 0755)
}

func profilePath(name string) (string, error) {
	dir, err := profileDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

// Save takes a snapshot of current listening ports and saves as a profile.
func Save(name string) (*Profile, error) {
	ports, err := port.ScanPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to scan ports: %w", err)
	}

	var entries []PortEntry
	seen := make(map[int]bool)
	for _, p := range ports {
		if seen[p.Port] {
			continue
		}
		seen[p.Port] = true
		entries = append(entries, PortEntry{
			Port:        p.Port,
			Proto:       p.Proto,
			ProcessName: p.ProcessName,
		})
	}

	profile := &Profile{
		Name:      name,
		Ports:     entries,
		CreatedAt: time.Now(),
	}

	path, err := profilePath(name)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}

	return profile, os.WriteFile(path, data, 0644)
}

// Load reads a profile from disk.
func Load(name string) (*Profile, error) {
	path, err := profilePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// List returns all saved profile names.
func List() ([]string, error) {
	dir, err := profileDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names, nil
}

// Delete removes a profile.
func Delete(name string) error {
	path, err := profilePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// Check verifies if all ports in a profile are currently active.
func Check(name string) ([]CheckResult, error) {
	profile, err := Load(name)
	if err != nil {
		return nil, err
	}

	currentPorts, err := port.ScanPorts()
	if err != nil {
		return nil, err
	}

	// Build lookup
	portMap := make(map[int]port.PortInfo)
	for _, p := range currentPorts {
		portMap[p.Port] = p
	}

	var results []CheckResult
	for _, entry := range profile.Ports {
		result := CheckResult{Entry: entry}
		if current, ok := portMap[entry.Port]; ok {
			result.Active = true
			result.Current = current.ProcessName
		}
		results = append(results, result)
	}

	return results, nil
}
