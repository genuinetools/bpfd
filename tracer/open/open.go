package open

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/genuinetools/bpfd/proc"
	"github.com/genuinetools/bpfd/tracer"
	bpf "github.com/iovisor/gobpf/bcc"
)

const (
	name = "open"
	// Mostly taken from: https://github.com/iovisor/bcc/blob/master/tools/opensnoop.py
	source string = `
#include <uapi/linux/ptrace.h>
#include <uapi/linux/limits.h>
#include <linux/sched.h>

typedef struct {
    u32 pid;  // PID as in the userspace term (i.e. task->tgid in kernel)
    u32 tgid; // Parent PID as in the userspace term (i.e task->real_parent->tgid in kernel)
    int ret;
    char comm[TASK_COMM_LEN];
    char filename[NAME_MAX];
} data_t;

BPF_HASH(tmp, u64, data_t);
BPF_PERF_OUTPUT(events);

int trace_entry(struct pt_regs *ctx, int dfd, const char __user *filename)
{
	u64 pid = bpf_get_current_pid_tgid();

	data_t data = {
		.pid = pid >> 32,
	};

    // Some kernels, like Ubuntu 4.13.0-generic, return 0
    // as the real_parent->tgid.
    // We use the get_tgid function as a fallback in those cases. (#1883)
    struct task_struct *task;
    task = (struct task_struct *)bpf_get_current_task();
    data.tgid = task->real_parent->tgid;

    bpf_get_current_comm(&data.comm, sizeof(data.comm));
	bpf_probe_read(&data.filename, sizeof(data.filename), (void *)filename);

    tmp.update(&pid, &data);
    return 0;
};
int trace_return(struct pt_regs *ctx)
{
    u64 id = bpf_get_current_pid_tgid();
    data_t *datap = tmp.lookup(&id);

    if (datap == 0) {
        // missed entry
        return 0;
    }

	data_t data = *datap;

    data.ret = PT_REGS_RC(ctx);
    events.perf_submit(ctx, &data, sizeof(data));

    tmp.delete(&id);
    return 0;
}
`
)

type openEvent struct {
	PID         uint32
	TGID        uint32
	ReturnValue int32
	Comm        [16]byte
	Filename    [255]byte
}

func init() {
	tracer.Register(name, Init)
}

type bpftracer struct {
	module  *bpf.Module
	perfMap *bpf.PerfMap
	channel chan []byte
}

// Init returns a new open tracer.
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

	openKprobe, err := p.module.LoadKprobe("trace_entry")
	if err != nil {
		return fmt.Errorf("load sys_open kprobe failed: %v", err)
	}

	open := "do_sys_open"
	err = p.module.AttachKprobe(open, openKprobe, -1)
	if err != nil {
		return fmt.Errorf("attach sys_open kprobe: %v", err)
	}

	openKretprobe, err := p.module.LoadKprobe("trace_return")
	if err != nil {
		return fmt.Errorf("load sys_open kretprobe failed: %v", err)
	}

	err = p.module.AttachKretprobe(open, openKretprobe, -1)
	if err != nil {
		return fmt.Errorf("attach sys_open kretprobe: %v", err)
	}

	table := bpf.NewTable(p.module.TableId("events"), p.module)

	p.perfMap, err = bpf.InitPerfMap(table, p.channel)
	if err != nil {
		return fmt.Errorf("init perf map failed: %v", err)
	}

	return nil
}

func (p *bpftracer) WatchEvent(ctx context.Context) (*grpc.Event, error) {
	var event openEvent
	data := <-p.channel
	if err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event); err != nil {
		return nil, fmt.Errorf("failed to decode received data: %v", err)
	}

	index := bytes.IndexByte(event.Filename[:], 0)
	if index <= -1 {
		index = 255
	}
	filename := strings.TrimSpace(string(event.Filename[:index]))

	index = bytes.IndexByte(event.Comm[:], 0)
	if index <= -1 {
		index = 16
	}
	command := strings.TrimSpace(string(event.Comm[:index]))

	// Ignore files with our own PID or we will have an infinite loop.
	if strings.HasPrefix(filename, fmt.Sprintf("/proc/%d", int(event.PID))) ||
		strings.HasPrefix(filename, fmt.Sprintf("/proc/%d", int(event.TGID))) ||
		command == "bpfd" {
		return nil, nil
	}

	// Get the UID and GID.
	uid, gid, err := proc.GetUIDGID(int(event.TGID), int(event.PID))
	if err != nil {
		return nil, fmt.Errorf("getting uid and gid for process %d failed: %v", event.PID, err)
	}

	e := &grpc.Event{
		PID:         event.PID,
		TGID:        event.TGID,
		UID:         uid,
		GID:         gid,
		Command:     command,
		ReturnValue: event.ReturnValue,
		Data: map[string]string{
			"filename": filename,
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
