package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/fatih/color"
)

// ExposeLAN starts a reverse proxy on 0.0.0.0 to expose a local port to LAN.
func ExposeLAN(ctx context.Context, targetPort, listenPort int) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	dim := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)
	magenta := color.New(color.FgMagenta, color.Bold)

	targetURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			red := color.New(color.FgRed)
			red.Printf("  ✗ Proxy error: %v\n", err)
			w.WriteHeader(http.StatusBadGateway)
			io.WriteString(w, "502 Bad Gateway — target not reachable")
		},
	}

	listenAddr := fmt.Sprintf("0.0.0.0:%d", listenPort)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: proxy,
	}

	// Get LAN IPs
	ips := getLANIPs()

	fmt.Println()
	green.Println("  ✓ LAN Expose is live!")
	fmt.Println()
	cyan.Printf("  🔗 Forwarding → localhost:%d\n", targetPort)
	fmt.Println()

	magenta.Println("  📱 Access from any device on your network:")
	fmt.Println()
	for _, ip := range ips {
		fmt.Printf("    • http://%s:%d\n", ip, listenPort)
	}
	if len(ips) == 0 {
		fmt.Printf("    • http://0.0.0.0:%d\n", listenPort)
	}
	fmt.Println()

	dim.Println("  💡 Use on your phone, tablet, or other PC on same WiFi")
	dim.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	yellow.Println("\n  ⏹ LAN expose stopped.")
	fmt.Println()
	return nil
}

// getLANIPs returns all non-loopback IPv4 addresses.
func getLANIPs() []string {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ip := ipNet.IP
			if ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}

	return ips
}
