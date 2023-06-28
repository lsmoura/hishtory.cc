LDFLAGS=-s -w

build-client:
	go build -ldflags "$(LDFLAGS)" -o bin/hishtory cmd/hishtory/*.go

build-server:
	go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/*.go
