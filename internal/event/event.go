package event

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "time"
)

type Event struct {
    Pid        uint32
    Ppid       uint32
    ExitCode   uint32
	_          [4]byte 
    DurationNs uint64
    Comm       [16]byte
    Filename   [127]byte
    ExitEvent  uint8
}

func UnmarshalBinary(data []byte) (*Event, error) {
    var event Event
    reader := bytes.NewReader(data)
    if err := binary.Read(reader, binary.LittleEndian, &event); err != nil {
        return nil, err
    }
    return &event, nil
}

func PrintEventInfo(e *Event) {
    timestamp := time.Now().Format("15:04:05")
    comm := string(bytes.TrimRight(e.Comm[:], "\x00"))
    
    if e.ExitEvent != 0 {  // Kiểm tra != 0 thay vì chỉ e.ExitEvent
        fmt.Printf("%-8s %-5s %-16s %-7d %-7d [%d]",
            timestamp, "EXIT", comm, e.Pid, e.Ppid, e.ExitCode)
        if e.DurationNs > 0 {
            fmt.Printf(" (%dms)", e.DurationNs/1000000)
        }
        fmt.Println()
    } else {
        filename := string(bytes.TrimRight(e.Filename[:], "\x00"))
        fmt.Printf("%-8s %-5s %-16s %-7d %-7d %s\n",
            timestamp, "EXEC", comm, e.Pid, e.Ppid, filename)
    }
}