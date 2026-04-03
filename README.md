# 🚀 PortMan — Local Port Manager

A blazing fast CLI tool to **list**, **kill**, **investigate**, and **monitor** local ports.  
Built with Go. Optimized for Windows, macOS, and Linux. Perfect for microservice development.

---

## ⚡ Quick Install

### Option 1: Automatic Setup (Recommended for Windows)

1. Go to [**Releases**](https://github.com/hnq1707/portman/releases/latest)
2. Download `portman-windows-amd64.exe`
3. Just **run the EXE once**. PortMan will automatically:
   - Create a permanent installation folder.
   - Add itself to your system **PATH**.
   - Rename itself to a clean `portman.exe`.
4. Open a **New Terminal** and type: `portman`

### Option 2: `go install` (For developers with Go installed)

```bash
go install github.com/hnq1707/portman@latest
```

### Option 3: Manual Install (macOS/Linux)

1. Download the binary for your OS.
2. `chmod +x portman-*`
3. Move to your bin path: `sudo mv portman-* /usr/local/bin/portman`

---

## 📖 Main Commands

### 📡 List ports
```bash
portman list              # All listening ports
portman list --port 8080  # Filter by port
portman list --filter node # Filter by process name
portman list --json       # JSON output for scripting
portman list --free 8000-9000  # Find available ports in range
```

### 💀 Kill processes
```bash
portman kill 8080           # Kill with confirmation
portman kill 3000 8080 5432 # Kill multiple ports
portman kill 8080 -f        # Force kill (skip prompt)
portman kill --all node     # Kill ALL ports used by 'node'
```

### 🔍 Deep Investigate
```bash
portman why 3000    # Process tree, memory, config files, related ports
```

### 👁 Real-time Monitoring
```bash
portman watch             # Live timeline: NEW/GONE/SWAP events
portman dashboard         # Interactive TUI dashboard with filter & kill
```

### 🏥 System Diagnostics
```bash
portman doctor        # Health score, conflict detection, permission checks
portman doctor --fix  # Interactive auto-fix for conflicts
```

### 🔗 Networking & Tunnels
```bash
portman map 3000 8080              # Local forwarding :3000 → :8080
portman expose 3000                # Tunnel to internet via pinggy
portman expose 3000 --local        # Expose to Local Area Network (LAN)
```

---

## 🏗️ Build from Source

**Requirements:** Go 1.21+

```bash
# Clone
git clone https://github.com/hnq1707/portman.git
cd portman

# Build
go build -o portman.exe .
```

---

## 🎯 Feature Matrix

| Feature | Command | Description |
|---------|---------|-------------|
| 📡 List | `list` | Pretty table with labels for well-known ports |
| 💀 Kill | `kill` | Single/multi/by-name termination |
| 🔍 Why | `why` | Detailed process forensics & config scan |
| 👁 Watch | `watch` | Real-time diff timeline of port activity |
| 🖥️ TUI | `dashboard` | Full interactive dashboard with search & navigation |
| 🏥 Doctor | `doctor` | Health score, conflict detection, Admin check |
| 🔗 Proxy | `map` | Bidirectional TCP proxy with live stats |
| 🌐 Tunnel | `expose` | Public/LAN tunneling support |
| 🐳 Docker | `docker` | List Docker container port bindings |
| 🩹 Setup | `setup` | Automatic PATH configuration for Windows |

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI architecture |
| [fatih/color](https://github.com/fatih/color) | Terminal styling |
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | TUI styling |

## License

MIT - See [LICENSE](LICENSE) for details.
