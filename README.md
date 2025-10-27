# eBPF Process Monitor

This repository is an adaptation of the [eunomia-bpf bootstrap example](https://github.com/eunomia-bpf/bpf-developer-tutorial/blob/main/src/11-bootstrap/README.md), using Go with [cilium/ebpf](https://github.com/cilium/ebpf) library for the userspace code instead of the original C implementation.

## Overview

This project demonstrates process monitoring using eBPF by tracing process execution and exit events. It captures:
- Process execution (via `sched_process_exec` tracepoint)
- Process exit (via `sched_process_exit` tracepoint)
- Process metadata (PID, PPID, command name, filename, exit code, duration)

## Key Differences from Original

- **Userspace Language**: Go instead of C
- **BPF Library**: [cilium/ebpf](https://github.com/cilium/ebpf) instead of libbpf
- **Code Generation**: Uses `bpf2go` for generating Go bindings from eBPF C code
- **Modern API**: Leverages Go's type safety and error handling

## Architecture

```
ebpf-bootstrap/
├── bpf/
│   ├── bootstrap.bpf.c    # eBPF kernel-space program (C)
│   ├── bootstrap.h         # Event structure definitions
│   └── vmlinux.h          # Kernel type definitions
├── cmd/
│   └── ebpf-bootstrap.go  # Main entry point
├── internal/
│   ├── event/
│   │   └── event.go       # Event parsing and formatting
│   ├── probe/
│   │   ├── probe.go       # eBPF loader and manager
│   │   ├── probe_bpfeb.go # Generated (big-endian)
│   │   └── probe_bpfel.go # Generated (little-endian)
│   └── timer/
│       └── timer.go       # Timing utilities
└── Makefile
```

## Requirements

- Linux kernel 5.8+ with BTF support
- Go 1.21+
- `clang` and `llvm` for compiling eBPF programs
- Root/CAP_BPF privileges to load eBPF programs

## Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd ebpf-bootstrap

# Install dependencies
go mod download

# Generate eBPF bindings and build
make build
```

## Usage

```bash
# Run with default settings
sudo ./ebpf-bootstrap

# Set minimum process duration filter (in milliseconds)
sudo ./ebpf-bootstrap -d 100
```

### Command-line Options

- `-d <duration>`: Minimum process duration in milliseconds to report (default: 0)

## Output Format

```
TIME     EVENT COMM             PID     PPID    FILENAME/EXIT CODE
23:02:38 EXEC  bash             1234    1000    /usr/bin/bash
23:02:39 EXEC  ls               1235    1234    /usr/bin/ls
23:02:39 EXIT  ls               1235    1234    [0] (12ms)
```

- **TIME**: Event timestamp (HH:MM:SS)
- **EVENT**: EXEC (process started) or EXIT (process terminated)
- **COMM**: Process command name (truncated to 16 chars)
- **PID**: Process ID
- **PPID**: Parent Process ID
- **FILENAME/EXIT CODE**: Executable path (EXEC) or exit code with duration (EXIT)

## How It Works

### Kernel-Space (eBPF)

The eBPF programs (`bpf/bootstrap.bpf.c`) attach to kernel tracepoints:

1. **`handle_exec`**: Triggered when a process executes
   - Captures PID, PPID, command name, and executable path
   - Records timestamp for duration calculation

2. **`handle_exit`**: Triggered when a process exits
   - Retrieves start time from hash map
   - Calculates process duration
   - Filters based on minimum duration threshold
   - Captures exit code

Events are sent to userspace via a ring buffer.

### User-Space (Go)

The Go application:

1. Loads the compiled eBPF object into the kernel
2. Attaches programs to tracepoints
3. Reads events from the ring buffer
4. Parses and formats event data
5. Displays real-time process information

## Development

### Modifying eBPF Code

1. Edit `bpf/bootstrap.bpf.c`
2. Run `make generate` to regenerate Go bindings
3. Rebuild with `make build`

### Code Generation

The `//go:generate` directive in `internal/probe/probe.go` uses `bpf2go`:

```go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64,arm64 -type event probe ../../bpf/bootstrap.bpf.c -- -I../../bpf
```

This generates:
- `probe_bpfel.go` / `probe_bpfeb.go`: Architecture-specific bindings
- `probe_bpfel.o` / `probe_bpfeb.o`: Compiled eBPF bytecode

## Troubleshooting

**Error: Failed to load BPF object**
- Ensure you're running with root privileges
- Check kernel version supports BTF (`ls /sys/kernel/btf/vmlinux`)

**Error: Permission denied**
- Run with `sudo` or grant CAP_BPF capability
- Check `/sys/fs/bpf` is mounted

## References

- [Original eunomia-bpf bootstrap tutorial](https://github.com/eunomia-bpf/bpf-developer-tutorial/blob/main/src/11-bootstrap/README.md)
- [cilium/ebpf library](https://github.com/cilium/ebpf)
- [eBPF documentation](https://ebpf.io/)
- [libbpf bootstrap example](https://github.com/libbpf/libbpf-bootstrap)

## License

Dual BSD/GPL
