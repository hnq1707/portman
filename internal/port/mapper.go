package port

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

// Mapper handles TCP port forwarding.
type Mapper struct {
	SrcPort       int
	DstPort       int
	Host          string
	listener      net.Listener
	connCount     atomic.Int64
	activeConns   sync.WaitGroup
	TotalBytesIn  atomic.Int64
	TotalBytesOut atomic.Int64
	ActiveConns   atomic.Int64
	TotalConns    atomic.Int64
}

// Stats returns current traffic statistics.
func (m *Mapper) Stats() (bytesIn, bytesOut, active, total int64) {
	return m.TotalBytesIn.Load(), m.TotalBytesOut.Load(),
		m.ActiveConns.Load(), m.TotalConns.Load()
}

// NewMapper creates a new port mapper.
func NewMapper(srcPort, dstPort int, host string) *Mapper {
	return &Mapper{
		SrcPort: srcPort,
		DstPort: dstPort,
		Host:    host,
	}
}

// Start begins listening and forwarding connections.
func (m *Mapper) Start(ctx context.Context) error {
	listenAddr := fmt.Sprintf("%s:%d", m.Host, m.SrcPort)
	targetAddr := fmt.Sprintf("%s:%d", m.Host, m.DstPort)

	var err error
	m.listener, err = net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Printf("\n  🔗 Port mapping active: %s → %s\n", listenAddr, targetAddr)

	dim := color.New(color.FgHiBlack)
	dim.Println("  Press Ctrl+C to stop\n")

	// Close listener when context is cancelled
	go func() {
		<-ctx.Done()
		m.listener.Close()
	}()

	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				// Graceful shutdown
				m.activeConns.Wait()
				return nil
			default:
				return fmt.Errorf("accept error: %w", err)
			}
		}

		m.activeConns.Add(1)
		m.ActiveConns.Add(1)
		m.TotalConns.Add(1)
		connID := m.connCount.Add(1)
		go m.handleConn(ctx, conn, targetAddr, connID)
	}
}

func (m *Mapper) handleConn(ctx context.Context, src net.Conn, targetAddr string, connID int64) {
	defer m.activeConns.Done()
	defer m.ActiveConns.Add(-1)
	defer src.Close()

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	dim := color.New(color.FgHiBlack)

	start := time.Now()
	green.Printf("  ▶ [#%d] %s connected\n", connID, src.RemoteAddr())

	dst, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	if err != nil {
		red.Printf("  ✗ [#%d] failed to connect to %s: %v\n", connID, targetAddr, err)
		return
	}
	defer dst.Close()

	// Bidirectional copy
	done := make(chan struct{})
	var bytesIn, bytesOut int64

	go func() {
		bytesOut, _ = io.Copy(dst, src)
		dst.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()

	go func() {
		bytesIn, _ = io.Copy(src, dst)
		src.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()

	// Wait for both directions to finish or context to cancel
	select {
	case <-done:
		<-done // Wait for the other direction too
	case <-ctx.Done():
	}

	// Track totals
	m.TotalBytesOut.Add(bytesOut)
	m.TotalBytesIn.Add(bytesIn)

	elapsed := time.Since(start)
	dim.Printf("  ◀ [#%d] closed (%s, ↑%s ↓%s)\n",
		connID, elapsed.Round(time.Millisecond),
		formatBytes(bytesOut), formatBytes(bytesIn))
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
