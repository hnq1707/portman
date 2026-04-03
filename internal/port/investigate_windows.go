//go:build windows

package port

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type psProcessResult struct {
	Name        string  `json:"Name"`
	CommandLine string  `json:"CommandLine"`
	ExePath     string  `json:"ExePath"`
	ParentPID   int     `json:"ParentPID"`
	MemoryMB    float64 `json:"MemoryMB"`
	StartTime   string  `json:"StartTime"`
	User        string  `json:"User"`
	Tree        []struct {
		PID       int    `json:"PID"`
		Name      string `json:"Name"`
		ParentPID int    `json:"ParentPID"`
	} `json:"Tree"`
}

// InvestigatePort performs a deep investigation of what's using a port on Windows.
func InvestigatePort(portNum int) (*ProcessDetail, error) {
	// 1. Find process on this port
	entries, err := FindByPort(portNum)
	if err != nil {
		return nil, fmt.Errorf("failed to scan port %d: %w", portNum, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no process found on port %d — port is free", portNum)
	}

	entry := entries[0]
	detail := &ProcessDetail{
		Port:        portNum,
		PID:         entry.PID,
		ProcessName: entry.ProcessName,
	}

	// 2. Get detailed process info via a SINGLE PowerShell call
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$targetPID = %d
$p = Get-CimInstance Win32_Process -Filter "ProcessId = $targetPID"
if (-not $p) { Write-Output '{}'; exit }

$pp = Get-Process -Id $targetPID -ErrorAction SilentlyContinue

# Memory in MB
$memMB = 0
if ($pp) { $memMB = [math]::Round($pp.WorkingSet64/1MB, 1) }

# Start time as ISO string
$startStr = ''
if ($p.CreationDate) { $startStr = $p.CreationDate.ToString('o') }

# User (may need elevation)
$user = ''
try { $user = (Get-Process -Id $targetPID -IncludeUserName -ErrorAction Stop).UserName } catch {}

# Build process tree — walk up parent chain
$chain = @()
$pid = $targetPID
$seen = @{}
while ($pid -gt 0 -and -not $seen.ContainsKey($pid) -and $chain.Count -lt 10) {
    $seen[$pid] = $true
    $cp = Get-CimInstance Win32_Process -Filter "ProcessId = $pid" -ErrorAction SilentlyContinue
    if (-not $cp) { break }
    $chain += [PSCustomObject]@{
        PID = [int]$cp.ProcessId
        Name = [string]$cp.Name
        ParentPID = [int]$cp.ParentProcessId
    }
    $pid = $cp.ParentProcessId
}

[PSCustomObject]@{
    Name = [string]$p.Name
    CommandLine = [string]$p.CommandLine
    ExePath = [string]$p.ExecutablePath
    ParentPID = [int]$p.ParentProcessId
    MemoryMB = $memMB
    StartTime = $startStr
    User = $user
    Tree = $chain
} | ConvertTo-Json -Depth 3
`, entry.PID)

	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err == nil {
		var result psProcessResult
		if json.Unmarshal(out, &result) == nil {
			detail.CommandLine = result.CommandLine
			detail.ExecutablePath = result.ExePath
			detail.ParentPID = result.ParentPID
			detail.MemoryMB = result.MemoryMB
			detail.User = result.User

			// Parse start time
			if result.StartTime != "" {
				if t, err := time.Parse(time.RFC3339, result.StartTime); err == nil {
					detail.StartTime = t
				}
			}

			// Process tree (PowerShell returns bottom-up, we reverse to top-down)
			for i := len(result.Tree) - 1; i >= 0; i-- {
				n := result.Tree[i]
				detail.ProcessTree = append(detail.ProcessTree, TreeNode{
					PID:       n.PID,
					Name:      n.Name,
					Depth:     len(result.Tree) - 1 - i,
					IsCurrent: n.PID == entry.PID,
				})
			}
		}
	}

	// 3. Infer project directory
	detail.WorkingDir = inferProjectDir(detail.CommandLine, detail.ExecutablePath)

	// 4. Find related ports (same PID, different port)
	allPorts, _ := ScanPorts()
	for _, p := range allPorts {
		if p.PID == entry.PID && p.Port != portNum {
			detail.RelatedPorts = append(detail.RelatedPorts, p)
		}
	}

	// 5. Scan config files in project directory
	if detail.WorkingDir != "" {
		detail.ConfigFiles = scanConfigFiles(detail.WorkingDir, portNum)
	}

	return detail, nil
}
