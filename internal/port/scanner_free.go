package port

import (
	"fmt"
	"net"
	"sync"
)

// FreePortResult holds info about a scanned port.
type FreePortResult struct {
	Port int  `json:"port"`
	Free bool `json:"free"`
}

// FindFreePorts scans a range and returns free ports.
// If count > 0, stops after finding that many free ports.
func FindFreePorts(start, end, count int) []int {
	if start > end {
		start, end = end, start
	}

	var mu sync.Mutex
	var freePorts []int
	sem := make(chan struct{}, 100) // concurrency limit
	var wg sync.WaitGroup

	for p := start; p <= end; p++ {
		if count > 0 {
			mu.Lock()
			done := len(freePorts) >= count
			mu.Unlock()
			if done {
				break
			}
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(port int) {
			defer wg.Done()
			defer func() { <-sem }()

			if IsPortFree(port) {
				mu.Lock()
				if count <= 0 || len(freePorts) < count {
					freePorts = append(freePorts, port)
				}
				mu.Unlock()
			}
		}(p)
	}

	wg.Wait()
	return freePorts
}

// IsPortFree checks if a port is available to bind.
func IsPortFree(p int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
