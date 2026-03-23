.PHONY: build install clean

# Build binary
build:
	go build -ldflags="-s -w" -o portman.exe .

# Install to GOPATH/bin (available globally)
install:
	go install -ldflags="-s -w" .

# Clean build artifacts
clean:
	rm -f portman.exe portman

# Build for all platforms
release:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/portman-windows-amd64.exe .
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/portman-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/portman-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/portman-darwin-arm64 .
