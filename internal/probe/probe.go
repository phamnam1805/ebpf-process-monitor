//go:build linux

package probe

import (
	"context"
	"log"
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"golang.org/x/sys/unix"

	"ebpf-bootstrap/internal/event"
)
//go:generate env GOPACKAGE=probe go run github.com/cilium/ebpf/cmd/bpf2go probe ../../bpf/bootstrap.bpf.c -- -O2

const tenMegaBytes = 1024 * 1024 * 10
const twentyMegaBytes = tenMegaBytes * 2
const fortyMegaBytes = twentyMegaBytes * 2

type probe struct {
	bpfObjects *probeObjects
	handleExecLink link.Link
	handleExitLink link.Link
}

func htons(hostOrder uint16) uint16 {
    return (hostOrder << 8) | (hostOrder >> 8)
}

func htonl(hostOrder uint32) uint32 {
    return ((hostOrder & 0xFF) << 24) |
           (((hostOrder >> 8) & 0xFF) << 16) |
           (((hostOrder >> 16) & 0xFF) << 8) |
           ((hostOrder >> 24) & 0xFF)
}

func setRlimit() error {
     log.Println("Setting rlimit")

     return unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
         Cur: twentyMegaBytes,
         Max: fortyMegaBytes,
     })
}

func (p *probe) loadObjects(minDuration int) error {
	log.Printf("Loading probe object into kernel")

	objs := probeObjects{}

	spec, err := loadProbe()
	if err != nil {
		return err
	}

	if minDuration > 0 {
		if err := spec.Variables["min_duration_ns"].Set(uint64(minDuration * 1e6)); err != nil {
			log.Printf("Failed setting min_duration_ns: %v", err)
			return err
		}
		// if err := spec.RewriteConstants(map[string]interface{}{
		// 	"min_duration_ns": uint64(minDuration * 1e6),
		// }); err != nil {
		// 	log.Printf("Failed rewriting constants: %v", err)
		// 	return err
		// }

		log.Printf("Set min_duration_ns to %d ms", minDuration)
	}
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}); err != nil {
		log.Printf("Failed loading probe objects: %v", err)
		return err
	}

	p.bpfObjects = &objs

	return nil
}


func (p *probe) attachPrograms() error {
	log.Printf("Attaching bpf programs to kernel")

	handleExecLink, err := link.Tracepoint("sched", "sched_process_exec", p.bpfObjects.HandleExec, nil)
    if err != nil {
        log.Printf("Failed to link tracepoint tp/sched/sched_process_exec: %v", err)
        return err
    }
    p.handleExecLink = handleExecLink

	log.Printf("Successfully linked tracepoint tp/sched/sched_process_exec")

	handleExitLink, err := link.Tracepoint("sched", "sched_process_exit", p.bpfObjects.HandleExit, nil)
    if err != nil {
        log.Printf("Failed to link tracepoint tp/sched/sched_process_exit: %v", err)
        return err
    }
    p.handleExitLink = handleExitLink

	log.Printf("Successfully linked tracepoint tp/sched/sched_process_exit")

	return nil
}

func newProbe(minDuration int) (*probe, error) {
	log.Println("Creating a new probe")


	prbe := probe{
	}

	if err := prbe.loadObjects(minDuration); err != nil {
		log.Printf("Failed loading probe objects: %v", err)
		return nil, err
	}

	if err := prbe.attachPrograms(); err != nil {
		log.Printf("Failed attaching bpf programs: %v", err)
		return nil, err
	}

	return &prbe, nil
}


func (p *probe) Close() error {
	log.Println("Closing eBPF object")

	if p.handleExecLink != nil {
        p.handleExecLink.Close()
    }

	if p.handleExitLink != nil {
        p.handleExitLink.Close()
    }

	return nil
}

func Run(ctx context.Context, minDuration int) error {
	log.Println("Starting up the probe")

	if err := setRlimit(); err != nil {
		log.Printf("Failed setting rlimit: %v", err)
		return err
	}

	probe, err := newProbe(minDuration)
	if err != nil {
		log.Printf("Failed creating new probe: %v", err)
		return err
	}
	
	eventPipe := probe.bpfObjects.probeMaps.Rb

	eventReader, err := ringbuf.NewReader(eventPipe)
	if err != nil {
		log.Fatalf("opening ringbuf reader: %s", err)
	}
	defer eventReader.Close()

	fmt.Printf("%-8s %-5s %-16s %-7s %-7s %s\n",
        "TIME", "EVENT", "COMM", "PID", "PPID", "FILENAME/EXIT CODE")

	go func() {
        for {
            record, err := eventReader.Read()
            if err != nil {
                if ctx.Err() != nil {
                    return
                }
                log.Printf("Failed reading from ringbuf: %v", err)
                continue
            }
            eventAttrs, err := event.UnmarshalBinary(record.RawSample)
            if err != nil {
                log.Printf("Could not unmarshal event: %+v", record.RawSample)
                continue
            }
            event.PrintEventInfo(eventAttrs)
        }
    }()

	<-ctx.Done()
    log.Println("Context cancelled, shutting down...")
    return probe.Close()
}