package port

// PortInfo holds information about a port binding.
type PortInfo struct {
	Proto       string `json:"proto"`
	LocalAddr   string `json:"local_addr"`
	Port        int    `json:"port"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	State       string `json:"state"`
}

// KillResult holds the result of a kill operation.
type KillResult struct {
	Port        int
	PID         int
	ProcessName string
	Success     bool
	Error       string
}

// FindByPort returns port info entries matching the given port number.
func FindByPort(port int) ([]PortInfo, error) {
	all, err := ScanPorts()
	if err != nil {
		return nil, err
	}

	var matched []PortInfo
	for _, p := range all {
		if p.Port == port {
			matched = append(matched, p)
		}
	}
	return matched, nil
}

// KillByPort finds the process using the given port and kills it.
func KillByPort(port int) []KillResult {
	entries, err := FindByPort(port)
	if err != nil {
		return []KillResult{{
			Port:    port,
			Success: false,
			Error:   "failed to scan ports: " + err.Error(),
		}}
	}

	if len(entries) == 0 {
		return []KillResult{{
			Port:    port,
			Success: false,
			Error:   "no process found on port",
		}}
	}

	// Deduplicate PIDs
	seen := make(map[int]bool)
	var results []KillResult

	for _, entry := range entries {
		if seen[entry.PID] {
			continue
		}
		seen[entry.PID] = true

		err := killPID(entry.PID)
		result := KillResult{
			Port:        port,
			PID:         entry.PID,
			ProcessName: entry.ProcessName,
		}
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}
		results = append(results, result)
	}

	return results
}
