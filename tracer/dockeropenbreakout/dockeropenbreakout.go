package dockeropenbreakout

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/genuinetools/bpfd/proc"
	"github.com/genuinetools/bpfd/tracer"
	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/sirupsen/logrus"
)

const (
	name = "dockeropenbreakout"
	// Mostly taken from: https://github.com/iovisor/bcc/blob/master/tools/opensnoop.py
	source string = `
#include <uapi/linux/ptrace.h>
#include <uapi/linux/limits.h>
#include <linux/sched.h>

typedef struct {
    u32 pid;  // PID as in the userspace term (i.e. task->tgid in kernel)
    u32 tgid; // Parent PID as in the userspace term (i.e task->real_parent->tgid in kernel)
	u32 uid;
	u32 gid;
    int ret;
    char comm[TASK_COMM_LEN];
    char filename[NAME_MAX];
} data_t;

BPF_HASH(tmp, u64, data_t);
BPF_PERF_OUTPUT(events);

int trace_entry(struct pt_regs *ctx, int dfd, const char __user *filename)
{
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
	UID         uint32
	GID         uint32
	ReturnValue int32
	Comm        [16]byte
	Filename    [255]byte
}

func init() {
	tracer.Register(name, Init)
}

type bpftracer struct {
	module       *bpf.Module
	perfMap      *bpf.PerfMap
	channel      chan []byte
	dockerClient *client.Client
}

// Init returns a new bashreadline tracer.
func Init() (tracer.Tracer, error) {
	// Create the docker client.
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("creating docker client from env failed: %v", err)
	}

	return &bpftracer{
		channel:      make(chan []byte),
		dockerClient: c,
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

	e := &grpc.Event{
		PID:  event.PID,
		TGID: event.TGID,
		UID:  event.UID,
		GID:  event.GID,
		Data: map[string]string{
			"filename":  filename,
			"command":   command,
			"returnval": fmt.Sprintf("%d", event.ReturnValue),
		}}

	// Only include events from docker runtime.
	e.ContainerRuntime = string(proc.GetContainerRuntime(int(event.TGID), int(event.PID)))
	if e.ContainerRuntime != string(proc.RuntimeDocker) {
		return nil, nil
	}

	// Get the container ID.
	e.ContainerID = proc.GetContainerID(int(event.TGID), int(event.PID))
	if len(e.ContainerID) < 1 {
		return nil, nil
	}

	// Get information for the container mounts.
	r, err := p.dockerClient.ContainerInspect(ctx, e.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("getting container inspect information for %s container id %s failed: %v", e.ContainerRuntime, e.ContainerID, err)
	}
	// Collect all the mount information from the graph driver and actual mounts.
	mounts := []string{}
	// Data looks like:
	// "Data": {
	//        "LowerDir": "/var/lib/docker/overlay/7efc3ec24e158ef58ce4103b079b8dda6b6fbccf005e1f08dc63817a70340b0b/root",
	//        "MergedDir": "/var/lib/docker/overlay/3606c689c896a921d4e076fd02a6743327767b1b3639ebc2bc8b6932536aa2c9/merged",
	//        "UpperDir": "/var/lib/docker/overlay/3606c689c896a921d4e076fd02a6743327767b1b3639ebc2bc8b6932536aa2c9/upper",
	//        "WorkDir": "/var/lib/docker/overlay/3606c689c896a921d4e076fd02a6743327767b1b3639ebc2bc8b6932536aa2c9/work"
	//    },
	// So iterate over it and add it to our mounts.
	for _, v := range r.GraphDriver.Data {
		mounts = append(mounts, v)
	}

	if hasPrefix(filename, mounts) {
		// The file is within the mount context of the container so return nil.
		logrus.Warnf("%s is in mounts: %#v", filename, mounts)
		return nil, nil
	}

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

func hasPrefix(str string, ss []string) bool {
	for _, s := range ss {
		if strings.HasPrefix(str, s) {
			return true
		}
	}

	return false
}
