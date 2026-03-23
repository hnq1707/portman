# 🚀 PortMan — Local Port Manager

A blazing fast CLI tool to **list**, **kill**, and **forward** local ports.  
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
