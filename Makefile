.PHONY: build test install

# Build the ffs binary
build:
	go build -o ffs ffs.go

# Run tests
test:
	go test -v ./...

# Install ffs binary to ~/.local/bin
install: build
	mkdir -p ~/.local/bin
	cp ffs ~/.local/bin

# Clean up
clean:
	rm -f ffs
