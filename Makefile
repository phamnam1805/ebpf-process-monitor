generate: 
	go generate ./...

build-ebpf-bootstrap:
	go build -ldflags "-s -w" -o ebpf-bootstrap cmd/ebpf-bootstrap.go

build: generate build-ebpf-bootstrap

clean:
	rm -f ebpf-bootstrap
	rm -f internal/probe/probe_bpf*.go
	rm -f internal/probe/probe_bpf*.o