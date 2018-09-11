package open

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/jessfraz/bpfd/program"
)

const (
	name = "open"
	// Mostly taken from: https://github.com/iovisor/bcc/blob/master/tools/opensnoop.py
	source string = `
#include <uapi/linux/ptrace.h>
#include <uapi/linux/limits.h>
#include <linux/sched.h>
struct val_t {
    u32 pid;  // PID as in the userspace term (i.e. task->tgid in kernel)
    u32 tgid; // Parent PID as in the userspace term (i.e task->real_parent->tgid in kernel)
    u64 timestamp;
    char comm[TASK_COMM_LEN];
    const char *filename;
};
struct data_t {
    u32 pid;  // PID as in the userspace term (i.e. task->tgid in kernel)
    u32 tgid; // Parent PID as in the userspace term (i.e task->real_parent->tgid in kernel)
    u64 timestamp;
    int ret;
    char comm[TASK_COMM_LEN];
    char filename[NAME_MAX];
};

BPF_HASH(infotmp, u64, struct val_t);
BPF_PERF_OUTPUT(events);

int trace_entry(struct pt_regs *ctx, int dfd, const char __user *filename)
{
    struct val_t val = {};
    struct task_struct *task;
	u64 pid = bpf_get_current_pid_tgid();
    val.pid = pid >> 32;
    task = (struct task_struct *)bpf_get_current_task();
    // Some kernels, like Ubuntu 4.13.0-generic, return 0
    // as the real_parent->tgid.
    // We use the get_tgid function as a fallback in those cases. (#1883)
    val.tgid = task->real_parent->tgid;
    bpf_get_current_comm(&val.comm, sizeof(val.comm));
    val.timestamp = bpf_ktime_get_ns();
    val.filename = filename;
    infotmp.update(&pid, &val);
    return 0;
};
int trace_return(struct pt_regs *ctx)
{
    u64 id = bpf_get_current_pid_tgid();
    struct val_t *valp;
    struct data_t data = {};
    u64 time = bpf_ktime_get_ns();
    valp = infotmp.lookup(&id);
    if (valp == 0) {
        // missed entry
        return 0;
    }
    bpf_probe_read(&data.comm, sizeof(data.comm), valp->comm);
    bpf_probe_read(&data.filename, sizeof(data.filename), (void *)valp->filename);
    data.pid = valp->pid;
    data.tgid = valp->tgid;
    data.timestamp = time / 1000;
    data.ret = PT_REGS_RC(ctx);
    events.perf_submit(ctx, &data, sizeof(data));
    infotmp.delete(&id);
    return 0;
}
`
)

type openEvent struct {
	PID         uint32
	TGID        uint32
	Timestamp   uint64
	ReturnValue int32
	Comm        [16]byte
	Filename    [255]byte
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

	openKprobe, err := p.module.LoadKprobe("trace_entry")
	if err != nil {
		return fmt.Errorf("load sys_open kprobe failed: %v", err)
	}

	open := "do_sys_open"
	err = p.module.AttachKprobe(open, openKprobe)
	if err != nil {
		return fmt.Errorf("attach sys_open kprobe: %v", err)
	}

	openKretprobe, err := p.module.LoadKprobe("trace_return")
	if err != nil {
		return fmt.Errorf("load sys_open kretprobe failed: %v", err)
	}

	err = p.module.AttachKretprobe(open, openKretprobe)
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

func (p *bpfprogram) WatchEvent() (*program.Event, error) {
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

	e := &program.Event{
		PID:  event.PID,
		TGID: event.TGID,
		Data: map[string]string{
			"filename":  filename,
			"command":   command,
			"returnval": fmt.Sprintf("%d", event.ReturnValue),
		}}

	return e, nil
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
