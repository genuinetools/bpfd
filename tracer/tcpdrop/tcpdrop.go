package tcpdrop

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/tcp"
	"github.com/jessfraz/bpfd/tracer"
)

const (
	name = "tcpdrop"

	source string = `
#include <uapi/linux/ptrace.h>
#include <uapi/linux/tcp.h>
#include <uapi/linux/ip.h>
#include <net/sock.h>
#include <bcc/proto.h>

typedef struct {
    u32 pid;  // PID as in the userspace term (i.e. task->tgid in kernel)
    u32 tgid; // Parent PID as in the userspace term (i.e task->real_parent->tgid in kernel)
    int ret;
    char comm[TASK_COMM_LEN];
	u32 saddr;
	u32 daddr;
	u16 sport;
	u16 dport;
	u8 state;
    u8 tcpflags;
} data_t;

BPF_PERF_OUTPUT(events);
BPF_HASH(tmp, u64, data_t);

static struct tcphdr *skb_to_tcphdr(const struct sk_buff *skb)
{
    // unstable API. verify logic in tcp_hdr() -> skb_transport_header().
    return (struct tcphdr *)(skb->head + skb->transport_header);
}

static inline struct iphdr *skb_to_iphdr(const struct sk_buff *skb)
{
    // unstable API. verify logic in ip_hdr() -> skb_network_header().
    return (struct iphdr *)(skb->head + skb->network_header);
}

// from include/net/tcp.h:
#ifndef tcp_flag_byte
#define tcp_flag_byte(th) (((u_int8_t *)th)[13])
#endif

/* Arguements from:
https://github.com/torvalds/linux/blob/7428b2e5d0b195f2a5e40f91d2b41a8503fcfe68/net/ipv4/tcp_input.c#L4396
*/
int trace_entry(struct pt_regs *ctx,
	struct sock *sk,
	struct sk_buff *skb)
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

	// pull in details from the packet headers and the sock struct
    u16 family = sk->__sk_common.skc_family;
    char state = sk->__sk_common.skc_state;
    u16 sport = 0, dport = 0;
    struct tcphdr *tcp = skb_to_tcphdr(skb);
    struct iphdr *ip = skb_to_iphdr(skb);
    u8 tcpflags = ((u_int8_t *)tcp)[13];
    sport = tcp->source;
    dport = tcp->dest;
    sport = ntohs(sport);
    dport = ntohs(dport);

	data.saddr = ip->saddr;
    data.daddr = ip->daddr;
    data.dport = dport;
    data.sport = sport;
    data.state = state;
    data.tcpflags = tcpflags;

	tmp.update(&pid, &data);
	return 0;
}
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

type tcpEvent struct {
	PID                uint32
	TGID               uint32
	ReturnValue        int32
	Comm               [16]byte
	SourceAddress      uint32
	DestinationAddress uint32
	SourcePort         uint16
	DestinationPort    uint16
	State              uint8
	TCPFlags           uint8
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

	tcpKprobe, err := p.module.LoadKprobe("trace_entry")
	if err != nil {
		return fmt.Errorf("load tcp_drop kprobe failed: %v", err)
	}

	tcp := "tcp_drop"
	err = p.module.AttachKprobe(tcp, tcpKprobe)
	if err != nil {
		return fmt.Errorf("attach tcp_drop kprobe: %v", err)
	}

	tcpKretprobe, err := p.module.LoadKprobe("trace_return")
	if err != nil {
		return fmt.Errorf("load tcp_drop kretprobe failed: %v", err)
	}

	err = p.module.AttachKretprobe(tcp, tcpKretprobe)
	if err != nil {
		return fmt.Errorf("attach tcp_drop kretprobe: %v", err)
	}

	table := bpf.NewTable(p.module.TableId("events"), p.module)

	p.perfMap, err = bpf.InitPerfMap(table, p.channel)
	if err != nil {
		return fmt.Errorf("init perf map failed: %v", err)
	}

	return nil
}

func (p *bpftracer) WatchEvent() (*grpc.Event, error) {
	var event tcpEvent
	data := <-p.channel
	if err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event); err != nil {
		return nil, fmt.Errorf("failed to decode received data: %v", err)
	}

	index := bytes.IndexByte(event.Comm[:], 0)
	if index <= -1 {
		index = 16
	}
	command := strings.TrimSpace(string(event.Comm[:index]))

	state, ok := tcp.States[event.State]
	if !ok {
		return nil, fmt.Errorf("%d is not a valid tcp state", event.State)
	}

	e := &grpc.Event{
		PID:  event.PID,
		TGID: event.TGID,
		Data: map[string]string{
			"saddr":     inetNtoa(event.SourceAddress),
			"daddr":     inetNtoa(event.DestinationAddress),
			"sport":     fmt.Sprintf("%d", event.SourcePort),
			"dport":     fmt.Sprintf("%d", event.DestinationPort),
			"state":     state,
			"tcpflags":  tcp.FlagsToString(event.TCPFlags),
			"command":   command,
			"returnval": fmt.Sprintf("%d", event.ReturnValue),
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

func inetNtoa(a uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(a), byte(a>>8), byte(a>>16), byte(a>>24))
}
