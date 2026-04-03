package watch

import (
	"sync"
	"time"

	"github.com/nay-kia/portman/internal/port"
)

// EventType represents a port change event type.
type EventType int

const (
	PortAppeared    EventType = iota // New port started listening
	PortDisappeared                  // Port stopped listening
	ProcessChanged                   // Same port, different process
)

// Event represents a single port change event.
type Event struct {
	Type      EventType
	Port      int
	Proto     string
	Process   string
	PID       int
	OldProc   string // only for ProcessChanged
	OldPID    int    // only for ProcessChanged
	Timestamp time.Time
}

// portState tracks what we last saw on a given port.
type portState struct {
	port.PortInfo
	LastSeen time.Time
}

// Watcher monitors ports and emits events when changes occur.
type Watcher struct {
	mu       sync.RWMutex
	previous map[string]portState // key: "proto:port"
	events   []Event
	filter   string
}

// NewWatcher creates a new port watcher.
func NewWatcher() *Watcher {
	return &Watcher{
		previous: make(map[string]portState),
	}
}

// portKey creates a unique key for a port entry.
func portKey(proto string, p int) string {
	return proto + ":" + string(rune('0'+p/10000)) + string(rune('0'+(p/1000)%10)) +
		string(rune('0'+(p/100)%10)) + string(rune('0'+(p/10)%10)) + string(rune('0'+p%10))
}

// Use fmt for proper key generation
func portKeyStr(proto string, p int) string {
	return proto + ":" + portIntToStr(p)
}

func portIntToStr(p int) string {
	if p == 0 {
		return "0"
	}
	result := make([]byte, 0, 5)
	for p > 0 {
		result = append([]byte{byte('0' + p%10)}, result...)
		p /= 10
	}
	return string(result)
}

// Poll scans current ports and returns new events since the last poll.
func (w *Watcher) Poll() ([]Event, []port.PortInfo, error) {
	current, err := port.ScanPorts()
	if err != nil {
		return nil, nil, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	var newEvents []Event

	// Build current map
	currentMap := make(map[string]port.PortInfo)
	for _, p := range current {
		key := portKeyStr(p.Proto, p.Port)
		currentMap[key] = p
	}

	// Check for disappeared or changed ports
	for key, prev := range w.previous {
		cur, exists := currentMap[key]
		if !exists {
			// Port disappeared
			newEvents = append(newEvents, Event{
				Type:      PortDisappeared,
				Port:      prev.Port,
				Proto:     prev.Proto,
				Process:   prev.ProcessName,
				PID:       prev.PID,
				Timestamp: now,
			})
			delete(w.previous, key)
		} else if cur.PID != prev.PID || cur.ProcessName != prev.ProcessName {
			// Process changed on same port
			newEvents = append(newEvents, Event{
				Type:      ProcessChanged,
				Port:      cur.Port,
				Proto:     cur.Proto,
				Process:   cur.ProcessName,
				PID:       cur.PID,
				OldProc:   prev.ProcessName,
				OldPID:    prev.PID,
				Timestamp: now,
			})
			w.previous[key] = portState{PortInfo: cur, LastSeen: now}
		} else {
			// Still the same — update last seen
			w.previous[key] = portState{PortInfo: cur, LastSeen: now}
		}
	}

	// Check for new ports
	for key, cur := range currentMap {
		if _, existed := w.previous[key]; !existed {
			newEvents = append(newEvents, Event{
				Type:      PortAppeared,
				Port:      cur.Port,
				Proto:     cur.Proto,
				Process:   cur.ProcessName,
				PID:       cur.PID,
				Timestamp: now,
			})
			w.previous[key] = portState{PortInfo: cur, LastSeen: now}
		}
	}

	// Store events
	w.events = append(w.events, newEvents...)

	// Cap event log at 200
	if len(w.events) > 200 {
		w.events = w.events[len(w.events)-200:]
	}

	return newEvents, current, nil
}

// AllEvents returns all recorded events.
func (w *Watcher) AllEvents() []Event {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]Event, len(w.events))
	copy(result, w.events)
	return result
}

// EventCount returns the number of events recorded.
func (w *Watcher) EventCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.events)
}
