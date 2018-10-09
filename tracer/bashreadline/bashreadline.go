package bashreadline

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/genuinetools/bpfd/tracer"
	bpf "github.com/iovisor/gobpf/bcc"
)

const (
	name = "bashreadline"
	// This is heavily based on: https://github.com/iovisor/gobpf/blob/master/examples/bcc/bash_readline/bash_readline.go
	source string = `
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

typedef struct {
	u32 pid;
	u32 tgid;
	u32 uid;
	u32 gid;
	char comm[80];
} data_t;

BPF_PERF_OUTPUT(events);

int get_return_value(struct pt_regs *ctx) {
	if (!PT_REGS_RC(ctx))
		return 0;

	u64 pid = bpf_get_current_pid_tgid();
	u64 uid = bpf_get_current_uid_gid();

	data_t data = {
		.pid = pid >> 32,
		.uid = uid & 0xffffffff,
		.gid = uid >> 32,
	};

    // Some kernels, like Ubuntu 4.13.0-generic, return 0
    // as the real_parent->tgid.
    // We use the get_tgid function as a fallback in those cases. (#1883)
    struct task_struct *task;
    task = (struct task_struct *)bpf_get_current_task();
    data.tgid = task->real_parent->tgid;

    bpf_get_current_comm(&data.comm, sizeof(data.comm));
    events.perf_submit(ctx, &data, sizeof(data));
    return 0;
}
`
)

type readlineEvent struct {
	PID  uint32
	TGID uint32
	UID  uint32
	GID  uint32
	Comm [80]byte
}

func init() {
	tracer.Register(name, Init)
}

type bpftracer struct {
	module  *bpf.Module
	perfMap *bpf.PerfMap
	channel chan []byte
}

// Init returns a new bashreadline tracer.
func Init() (tracer.Tracer, error) {
	return &bpftracer{
		channel: make(chan []byte),
	}, nil
}

func (p *bpftracer) String() string {
	return name
}

func (p *bpftracer) Load() error {
	p.module = bpf.NewModule(source, []string{})

	readlineUretprobe, err := p.module.LoadUprobe("get_return_value")
	if err != nil {
		return fmt.Errorf("load get_return_value uprobe failed: %v", err)
	}

	err = p.module.AttachUretprobe("/bin/bash", "readline", readlineUretprobe, -1)
	if err != nil {
		return fmt.Errorf("attach return_value ureturnprobe: %v", err)
	}

	table := bpf.NewTable(p.module.TableId("events"), p.module)

	p.perfMap, err = bpf.InitPerfMap(table, p.channel)
	if err != nil {
		return fmt.Errorf("init perf map failed: %v", err)
	}

	return nil
}

func (p *bpftracer) WatchEvent(ctx context.Context) (*grpc.Event, error) {
	var event readlineEvent
	data := <-p.channel
	err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode received data: %v", err)
	}

	// Convert C string (null-terminated) to Go string
	command := strings.TrimSpace(string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)]))

	e := &grpc.Event{
		PID:  event.PID,
		TGID: event.TGID,
		UID:  event.UID,
		GID:  event.GID,
		Data: map[string]string{
			"command": command,
		}}

	return e, nil
}

func (p *bpftracer) Start() {
	p.perfMap.Start()
}

func (p *bpftracer) Unload() {
	if p.perfMap != nil {
		p.perfMap.Stop()
	}
	if p.module != nil {
		p.module.Close()
	}
}
