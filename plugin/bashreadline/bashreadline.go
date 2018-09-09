// +build bashreadline

package bashreadline

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/jessfraz/bpfd/plugin"
	"github.com/sirupsen/logrus"
)

// This is heavily based on: https://github.com/iovisor/gobpf/blob/master/examples/bcc/bash_readline/bash_readline.go
const (
	name          = "bashreadline"
	source string = `
#include <uapi/linux/ptrace.h>
struct readline_event_t {
        u32 pid;
        char str[80];
} __attribute__((packed));
BPF_PERF_OUTPUT(readline_events);
int get_return_value(struct pt_regs *ctx) {
        struct readline_event_t event = {};
        u32 pid;
        if (!PT_REGS_RC(ctx))
                return 0;
        pid = bpf_get_current_pid_tgid();
        event.pid = pid;
        bpf_probe_read(&event.str, sizeof(event.str), (void *)PT_REGS_RC(ctx));
        readline_events.perf_submit(ctx, &event, sizeof(event));
        return 0;
}
`
)

type readlineEvent struct {
	Pid uint32
	Str [80]byte
}

func init() {
	plugin.Register(name, Init)
}

type program struct {
	module  *bpf.Module
	perfMap *bpf.PerfMap
	channel chan []byte
}

// Init returns a new bashreadline plugin.
func Init() (plugin.Plugin, error) {
	return &program{
		channel: make(chan []byte),
	}, nil
}

func (p *program) String() string {
	return name
}

func (p *program) Load() error {
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

func (p *program) WatchEvents() error {
	fmt.Printf("[bashreadline]: %10s\t%s\n", "PID", "COMMAND")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var event readlineEvent
		for {
			data := <-p.channel
			err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
			if err != nil {
				logrus.Errorf("failed to decode received data: %v", err)
				continue
			}

			// Convert C string (null-terminated) to Go string
			comm := string(event.Str[:bytes.IndexByte(event.Str[:], 0)])
			fmt.Printf("[bashreadline]: %10d\t%s\n", event.Pid, comm)
		}
	}()

	p.perfMap.Start()
	wg.Wait()
	p.perfMap.Stop()
	return nil
}

func (p *program) Unload() error {
	p.perfMap.Stop()
	p.module.Close()
	return nil
}
