# 🚀 PortMan — Local Port Manager

A blazing fast CLI tool to **list**, **kill**, and **forward** local ports.  
Built with Go. Perfect for microservice development.

---

## ⚡ Quick Install

### Option 1: `go install` (recommended)
```bash
go install github.com/nay-kia/portman@latest
```
> Binary sẽ nằm ở `$GOPATH/bin/portman` — dùng được ngay từ terminal.

### Option 2: Clone & Build
```bash
git clone https://github.com/nay-kia/portman.git
cd portman
go build -ldflags="-s -w" -o portman.exe .
```

### Option 3: Download Release
Tải binary từ [Releases](https://github.com/nay-kia/portman/releases), giải nén và thêm vào `PATH`.

---

## 📖 Usage

### 📡 List listening ports
```bash
portman list              # Tất cả ports đang listen
portman list --port 8080  # Filter theo port
portman list --json       # JSON output (pipe to jq)
```

### 💀 Kill process on a port
```bash
portman kill 8080           # Kill với confirmation
portman kill 3000 8080 5432 # Kill nhiều ports
portman kill 8080 -f        # Force kill (bỏ qua confirm)
```

### 🔗 Forward port traffic
```bash
portman map 3000 8080              # Forward :3000 → :8080
portman map 80 3000 --host 0.0.0.0 # Expose ra network
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

| Feature | Description |
|---------|-------------|
| 🎨 Colored output | Well-known ports (MySQL, Redis…) highlighted |
| ⚡ Fast | Native `netstat` parsing, zero overhead |
| 🔒 Safe | Confirmation before killing processes |
| 📊 JSON output | Pipe to `jq` for scripting |
| 🔗 Port forwarding | Bidirectional TCP proxy with stats |
| 🛑 Graceful shutdown | Ctrl+C cleanly stops forwarding |

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [fatih/color](https://github.com/fatih/color) | Terminal colors |

## License

MIT
