package doctor

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/hnq1707/portman/internal/port"
	"github.com/hnq1707/portman/internal/utils"
)

// Severity levels for diagnostic findings.
type Severity int

const (
	SeverityOK      Severity = iota // All good
	SeverityInfo                    // FYI
	SeverityWarning                 // Something to watch
	SeverityError                   // Needs attention
)

// Finding represents a single diagnostic result.
type Finding struct {
	Title       string
	Description string
	Severity    Severity
	Ports       []int
	PIDs        []int
	ProcessName string
	Fixable     bool
	FixAction   string // Human-readable fix description
}

// Report holds the complete diagnostic report.
type Report struct {
	Score    int // 0-100 health score
	Findings []Finding
	PortCount int
	ProcessCount int
}

// ProcessPortGroup groups ports by process for analysis.
type ProcessPortGroup struct {
	ProcessName string
	PID         int
	Ports       []port.PortInfo
}

// RunDiagnostics performs all health checks and returns a report.
func RunDiagnostics() (*Report, error) {
	ports, err := port.ScanPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to scan ports: %w", err)
	}

	report := &Report{
		Score:    100,
		PortCount: len(ports),
	}

	// Count unique processes
	pidSet := make(map[int]bool)
	for _, p := range ports {
		pidSet[p.PID] = true
	}
	report.ProcessCount = len(pidSet)

	// Group ports by process
	groups := groupByProcess(ports)

	// Run diagnostic checks
	report.Findings = append(report.Findings, checkPortConflicts(ports)...)
	report.Findings = append(report.Findings, checkPortHogging(groups)...)
	report.Findings = append(report.Findings, checkSuspiciousPorts(ports)...)
	report.Findings = append(report.Findings, checkHighPorts(ports)...)
	report.Findings = append(report.Findings, checkPrivilegedPorts(ports)...)
	report.Findings = append(report.Findings, checkAdminPrivileges()...)

	// Calculate score
	for _, f := range report.Findings {
		switch f.Severity {
		case SeverityError:
			report.Score -= 15
		case SeverityWarning:
			report.Score -= 5
		case SeverityInfo:
			report.Score -= 1
		}
	}
	if report.Score < 0 {
		report.Score = 0
	}

	return report, nil
}

func groupByProcess(ports []port.PortInfo) []ProcessPortGroup {
	pidMap := make(map[int]*ProcessPortGroup)
	for _, p := range ports {
		if group, ok := pidMap[p.PID]; ok {
			group.Ports = append(group.Ports, p)
		} else {
			pidMap[p.PID] = &ProcessPortGroup{
				ProcessName: p.ProcessName,
				PID:         p.PID,
				Ports:       []port.PortInfo{p},
			}
		}
	}

	var groups []ProcessPortGroup
	for _, g := range pidMap {
		groups = append(groups, *g)
	}

	// Sort by port count descending
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Ports) > len(groups[j].Ports)
	})

	return groups
}

// checkPortConflicts detects multiple different processes trying to use the same port.
func checkPortConflicts(ports []port.PortInfo) []Finding {
	portProcs := make(map[int][]port.PortInfo)
	for _, p := range ports {
		portProcs[p.Port] = append(portProcs[p.Port], p)
	}

	var findings []Finding
	for portNum, procs := range portProcs {
		if len(procs) <= 1 {
			continue
		}

		// Check if different PIDs are using the same port
		pidSet := make(map[int]bool)
		var names []string
		for _, p := range procs {
			if !pidSet[p.PID] {
				pidSet[p.PID] = true
				names = append(names, fmt.Sprintf("%s (PID %d)", p.ProcessName, p.PID))
			}
		}

		if len(pidSet) > 1 {
			findings = append(findings, Finding{
				Title:       fmt.Sprintf("Port conflict on :%d", portNum),
				Description: fmt.Sprintf("Multiple processes sharing port %d: %s", portNum, strings.Join(names, ", ")),
				Severity:    SeverityError,
				Ports:       []int{portNum},
				Fixable:     true,
				FixAction:   fmt.Sprintf("Kill conflicting processes on port %d", portNum),
			})
		}
	}

	return findings
}

// checkPortHogging detects processes using an unusually large number of ports.
func checkPortHogging(groups []ProcessPortGroup) []Finding {
	var findings []Finding
	threshold := 10

	for _, g := range groups {
		if len(g.Ports) >= threshold {
			portNums := make([]int, len(g.Ports))
			for i, p := range g.Ports {
				portNums[i] = p.Port
			}

			findings = append(findings, Finding{
				Title:       fmt.Sprintf("%s is using %d ports", g.ProcessName, len(g.Ports)),
				Description: fmt.Sprintf("Process %s (PID %d) has %d port bindings. This may indicate a port leak or misconfiguration.", g.ProcessName, g.PID, len(g.Ports)),
				Severity:    SeverityWarning,
				Ports:       portNums,
				PIDs:        []int{g.PID},
				ProcessName: g.ProcessName,
				Fixable:     false,
				FixAction:   "Review if all port bindings are intentional",
			})
		}
	}

	return findings
}

// checkSuspiciousPorts detects unknown processes on non-standard ports.
func checkSuspiciousPorts(ports []port.PortInfo) []Finding {
	var findings []Finding

	// System processes that are always expected
	systemProcs := map[string]bool{
		"system":        true,
		"svchost.exe":   true,
		"services.exe":  true,
		"lsass.exe":     true,
		"wininit.exe":   true,
		"csrss.exe":     true,
		"smss.exe":      true,
		"spoolsv.exe":   true,
		"systemd":       true,
		"init":          true,
		"sshd":          true,
	}

	// Common dev tools + desktop apps — not suspicious
	knownProcs := map[string]bool{
		"node.exe": true, "node": true,
		"java.exe": true, "java": true,
		"python.exe": true, "python": true, "python3": true,
		"go.exe": true, "go": true,
		"docker.exe": true, "docker": true,
		"postgres": true, "mysqld": true, "redis-server": true,
		"mongod": true, "nginx": true, "httpd": true,
		"code.exe": true, "code": true,
		// Desktop apps
		"chrome.exe": true, "chrome": true,
		"firefox.exe": true, "firefox": true,
		"msedge.exe": true,
		"spotify.exe": true, "spotify": true,
		"ms-teams.exe": true, "teams.exe": true,
		"slack.exe": true, "slack": true,
		"discord.exe": true, "discord": true,
		"onedrive.exe": true, "onedrive.sync.service.exe": true,
		"dropbox.exe": true,
		"antigravity.exe": true,
		"windowsterminal.exe": true,
		"explorer.exe": true,
	}

	// Deduplicate by port+PID
	type portPIDKey struct {
		port int
		pid  int
	}
	seen := make(map[portPIDKey]bool)

	for _, p := range ports {
		procLower := strings.ToLower(p.ProcessName)

		// Skip known system + dev + desktop processes
		if systemProcs[procLower] || knownProcs[procLower] {
			continue
		}

		// Skip known ports regardless of process
		if port.IsKnownPort(p.Port) {
			continue
		}

		// Deduplicate
		key := portPIDKey{p.Port, p.PID}
		if seen[key] {
			continue
		}
		seen[key] = true

		// Flag unknown process on unknown port
		findings = append(findings, Finding{
			Title:       fmt.Sprintf("Unknown service on :%d", p.Port),
			Description: fmt.Sprintf("Process '%s' (PID %d) is listening on port %d — not a recognized service", p.ProcessName, p.PID, p.Port),
			Severity:    SeverityInfo,
			Ports:       []int{p.Port},
			PIDs:        []int{p.PID},
			ProcessName: p.ProcessName,
			Fixable:     true,
			FixAction:   fmt.Sprintf("Investigate or kill '%s' on port %d", p.ProcessName, p.Port),
		})
	}

	return findings
}

// checkHighPorts detects services on very high ephemeral ports.
func checkHighPorts(ports []port.PortInfo) []Finding {
	var findings []Finding
	ephemeralPorts := 0

	for _, p := range ports {
		if p.Port >= 49152 {
			ephemeralPorts++
		}
	}

	if ephemeralPorts > 5 {
		findings = append(findings, Finding{
			Title:       fmt.Sprintf("%d services on ephemeral ports (49152+)", ephemeralPorts),
			Description: "Multiple services are binding to ephemeral port range. These ports are typically assigned dynamically by the OS.",
			Severity:    SeverityInfo,
			Fixable:     false,
		})
	}

	return findings
}

// checkPrivilegedPorts checks for non-system processes on privileged ports (< 1024).
func checkPrivilegedPorts(ports []port.PortInfo) []Finding {
	if runtime.GOOS == "windows" {
		return nil // Windows doesn't enforce privileged ports the same way
	}

	var findings []Finding
	for _, p := range ports {
		if p.Port < 1024 && !port.IsKnownPort(p.Port) {
			findings = append(findings, Finding{
				Title:       fmt.Sprintf("Privileged port :%d in use", p.Port),
				Description: fmt.Sprintf("Process '%s' is using privileged port %d (< 1024). This may require root/admin privileges.", p.ProcessName, p.PID),
				Severity:    SeverityWarning,
				Ports:       []int{p.Port},
				PIDs:        []int{p.PID},
				ProcessName: p.ProcessName,
				Fixable:     false,
			})
		}
	}

	return findings
}

// GetFixablePorts returns ports from findings that can be auto-fixed (killed).
func GetFixablePorts(findings []Finding) []int {
	seen := make(map[int]bool)
	var ports []int
	for _, f := range findings {
		if f.Fixable && f.Severity >= SeverityWarning {
			for _, p := range f.Ports {
				if !seen[p] {
					seen[p] = true
					ports = append(ports, p)
				}
			}
		}
	}
	return ports
}
func checkAdminPrivileges() []Finding {
	if !utils.IsAdmin() {
		return []Finding{
			{
				Title:       "Limited privileges",
				Description: "PortMan is not running as Administrator/Root. You may be unable to kill processes owned by system services or other users.",
				Severity:    SeverityWarning,
				Fixable:     false,
				FixAction:   "Restart terminal as Administrator/Root for full control",
			},
		}
	}
	return nil
}
