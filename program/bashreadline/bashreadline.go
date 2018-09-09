package bashreadline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/jessfraz/bpfd/program"
)

// This is heavily based on: https://github.com/iovisor/gobpf/blob/master/examples/bcc/bash_readline/bash_readline.go
const (
	name          = "bashreadline"
	source string = `
#include <uapi/linux/ptrace.h>
struct readline_event_t {
        u32 pid;
        char comm[80];
} __attribute__((packed));
BPF_PERF_OUTPUT(readline_events);
int get_return_value(struct pt_regs *ctx) {
        struct readline_event_t event = {};
        u32 pid;
        if (!PT_REGS_RC(ctx))
                return 0;
        pid = bpf_get_current_pid_tgid();
        event.pid = pid;
        bpf_probe_read(&event.comm, sizeof(event.comm), (void *)PT_REGS_RC(ctx));
        readline_events.perf_submit(ctx, &event, sizeof(event));
        return 0;
}
`
)

type readlineEvent struct {
	Pid  uint32
	Comm [80]byte
}

func init() {
	program.Register(name, Init)
}

type bpfprogram struct {
	module  *bpf.Module
	perfMap *bpf.PerfMap
	channel chan []byte
}

// Init returns a new bashreadline program.
func Init() (program.Program, error) {
	return &bpfprogram{
		channel: make(chan []byte),
	}, nil
}

func (p *bpfprogram) String() string {
	return name
}

func (p *bpfprogram) Load() error {
	p.module = bpf.NewModule(source, []string{})

	readlineUretprobe, err := p.module.LoadUprobe("get_return_value")
	if err != nil {
		return fmt.Errorf("load get_return_value uprobe failed: %v", err)
	}

	err = p.module.AttachUretprobe("/bin/bash", "readline", readlineUretprobe, -1)
	if err != nil {
		return fmt.Errorf("attach return_value ureturnprobe: %v", err)
	}

	table := bpf.NewTable(p.module.TableId("readline_events"), p.module)

	p.perfMap, err = bpf.InitPerfMap(table, p.channel)
	if err != nil {
		return fmt.Errorf("init perf map failed: %v", err)
	}

	return nil
}

func (p *bpfprogram) WatchEvent() (*program.Event, error) {
	var event readlineEvent
	data := <-p.channel
	err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode received data: %v", err)
	}

	// Convert C string (null-terminated) to Go string
	command := string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)])

	return &program.Event{PID: event.Pid, Data: map[string]string{"command": command}}, nil
}

func (p *bpfprogram) Start() {
	p.perfMap.Start()
}

func (p *bpfprogram) Unload() {
	if p.perfMap != nil {
		p.perfMap.Stop()
	}
	if p.module != nil {
		p.module.Close()
	}
}
