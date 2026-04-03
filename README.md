# 🚀 PortMan — Local Port Manager

A blazing fast CLI tool to **list**, **kill**, **investigate**, and **monitor** local ports.  
Built with Go. Perfect for microservice development.

---

## ⚡ Quick Install

### Option 1: Download binary (không cần cài gì)

1. Vào [**Releases**](https://github.com/nay-kia/portman/releases/latest)
2. Tải file phù hợp OS:
   | OS | File |
   |---|---|
   | Windows | `portman-windows-amd64.exe` |
   | macOS (Intel) | `portman-darwin-amd64` |
   | macOS (M1/M2) | `portman-darwin-arm64` |
   | Linux | `portman-linux-amd64` |
3. **Windows**: Tải về → chạy `install.bat` → xong! Mở terminal mới gõ `portman`
4. **macOS/Linux**: `chmod +x portman-*` → move vào `/usr/local/bin/`

### Option 2: `go install` (dành cho dev có Go)
```bash
go install github.com/nay-kia/portman@latest
```

### Option 3: Clone & Build
```bash
git clone https://github.com/nay-kia/portman.git
cd portman
go build -ldflags="-s -w" -o portman.exe .
```

---

## 📖 Usage

### 📡 List listening ports
```bash
portman list              # Tất cả ports đang listen
portman list --port 8080  # Filter theo port
portman list --filter node # Filter theo process name
portman list --dev        # Chỉ dev ports (3000-9999)
portman list --json       # JSON output (pipe to jq)
portman list --free 8000-9000  # Tìm ports trống trong range
```

### 💀 Kill process on a port
```bash
portman kill 8080           # Kill với confirmation
portman kill 3000 8080 5432 # Kill nhiều ports
portman kill 8080 -f        # Force kill (bỏ qua confirm)
portman kill --all node     # Kill tất cả ports của node
```

### 🔍 Deep investigate a port
```bash
portman why 3000    # Process tree, memory, config files, related ports
portman why 8080    # Interactive: kill, open dir
```

### 👁 Real-time port monitoring
```bash
portman watch             # Timeline view: NEW/GONE/SWAP events
portman watch --interval 5  # Custom poll interval
portman dashboard         # Full TUI dashboard with kill support
```

### 🏥 Port health diagnostics
```bash
portman doctor        # Health score, conflict detection, suspicious ports
portman doctor --fix  # Interactive auto-fix for issues
```

### 🔗 Forward port traffic
```bash
portman map 3000 8080              # Forward :3000 → :8080
portman map 80 3000 --host 0.0.0.0 # Expose ra network
```

### 🌐 Expose port to internet
```bash
portman expose 3000              # Tunnel via pinggy
portman expose 3000 --local      # LAN only (same WiFi)
portman expose 8080 --provider serveo  # Use serveo.net
```

### 🐳 Docker port bindings
```bash
portman docker    # List all Docker container port mappings
```

### 📋 Port profiles
```bash
portman profile save myapp       # Snapshot current ports
portman profile list             # List saved profiles
portman profile check myapp      # Verify profile ports are active
portman profile check            # Check .portman.yml in current dir
portman profile delete myapp     # Delete a profile
```

---

## 🏗️ Build from Source

**Requirements:** Go 1.21+

```bash
# Clone
git clone https://github.com/nay-kia/portman.git
cd portman

# Build
go build -ldflags="-s -w" -o portman.exe .

# Hoặc install vào GOPATH/bin
go install .

# Cross-platform build (cần make)
make release
```

### Build output:
```
dist/
├── portman-windows-amd64.exe
├── portman-linux-amd64
├── portman-darwin-amd64
└── portman-darwin-arm64
```

---

## 🎯 Features

| Feature | Command | Description |
|---------|---------|-------------|
| 📡 List | `list` | Pretty table with process info, well-known port labels |
| 💀 Kill | `kill` | Single/multi/by-name kill with confirmation |
| 🔍 Investigate | `why` | Process tree, memory, config scan, interactive actions |
| 👁 Watch | `watch` | Real-time port diff timeline (NEW/GONE/SWAP) |
| 🖥️ Dashboard | `dashboard` | Full TUI with filter, kill, auto-refresh |
| 🏥 Doctor | `doctor` | Health score, conflict detection, auto-fix |
| 🔗 Forward | `map` | Bidirectional TCP proxy with stats |
| 🌐 Expose | `expose` | Tunnel to internet or LAN |
| 🐳 Docker | `docker` | Docker container port bindings |
| 📋 Profile | `profile` | Save/load/check port snapshots |
| 📊 JSON | `list --json` | Pipe to jq for scripting |
| 🔍 Free scan | `list --free` | Find available ports in a range |

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [fatih/color](https://github.com/fatih/color) | Terminal colors |
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | TUI styling |

## License

MIT
