generate: 
	go generate ./...

build-ebpf-process-monitor:
	go build -ldflags "-s -w" -o ebpf-process-monitor cmd/ebpf-process-monitor.go

build: generate build-ebpf-process-monitor

clean:
	rm -f ebpf-process-monitor
	rm -f internal/probe/probe_bpf*.go
	rm -f internal/probe/probe_bpf*.o