package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ContainerInfo holds Docker container port info.
type ContainerInfo struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	Ports  []PortBinding `json:"ports"`
}

// PortBinding represents a Docker port binding.
type PortBinding struct {
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Proto         string `json:"proto"`
}

// dockerPS is the JSON format from docker ps.
type dockerPS struct {
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	Status string `json:"Status"`
	Ports  string `json:"Ports"`
}

// ListContainerPorts returns port bindings from running Docker containers.
func ListContainerPorts() ([]ContainerInfo, error) {
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("Docker not found. Install Docker to use this feature")
	}

	out, err := exec.Command("docker", "ps", "--format", "{{json .}}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'docker ps': %w", err)
	}

	outStr := strings.TrimSpace(string(out))
	if outStr == "" {
		return nil, nil
	}

	var containers []ContainerInfo
	lines := strings.Split(outStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var ps dockerPS
		if err := json.Unmarshal([]byte(line), &ps); err != nil {
			continue
		}

		container := ContainerInfo{
			Name:   ps.Names,
			Image:  ps.Image,
			Status: ps.Status,
			Ports:  parsePorts(ps.Ports),
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// parsePorts parses Docker port strings like "0.0.0.0:8080->80/tcp, 3306/tcp"
func parsePorts(portsStr string) []PortBinding {
	if portsStr == "" {
		return nil
	}

	var bindings []PortBinding
	parts := strings.Split(portsStr, ", ")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		binding := PortBinding{Proto: "tcp"}

		// Extract protocol
		if slashIdx := strings.LastIndex(part, "/"); slashIdx != -1 {
			binding.Proto = part[slashIdx+1:]
			part = part[:slashIdx]
		}

		// Check if there's a mapping (->)
		if arrowIdx := strings.Index(part, "->"); arrowIdx != -1 {
			hostPart := part[:arrowIdx]
			containerPart := part[arrowIdx+2:]

			// Host part can be "0.0.0.0:8080" or just "8080"
			if colonIdx := strings.LastIndex(hostPart, ":"); colonIdx != -1 {
				binding.HostPort = hostPart[colonIdx+1:]
			} else {
				binding.HostPort = hostPart
			}

			binding.ContainerPort = containerPart
		} else {
			// No mapping, just exposed port
			binding.ContainerPort = part
			binding.HostPort = "-"
		}

		bindings = append(bindings, binding)
	}

	return bindings
}
