package stdout

import (
	"fmt"

	"github.com/genuinetools/bpfd/action"
	"github.com/genuinetools/bpfd/api/grpc"
)

const (
	name = "stdout"
)

type stdoutAction struct{}

func init() {
	action.Register(name, Init)
}

// Init returns a new stdout action.
func Init() (action.Action, error) {
	return &stdoutAction{}, nil
}

func (s *stdoutAction) String() string {
	return name
}

func (s *stdoutAction) Do(event *grpc.Event) error {
	/*	logrus.WithFields(logrus.Fields{
		"tracer":            event.Tracer,
		"pid":               fmt.Sprintf("%d", event.PID),
		"tgid":              fmt.Sprintf("%d", event.TGID),
		"uid":               fmt.Sprintf("%d", event.UID),
		"gid":               fmt.Sprintf("%d", event.GID),
		"command":           event.Command,
		"return_value":      fmt.Sprintf("%d", event.ReturnValue),
		"container_runtime": string(event.ContainerRuntime),
		"container_id":      event.ContainerID,
	}).Infof("%#v", event.Data)*/

	fmt.Printf("%s\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
		event.Tracer,
		event.PID,
		event.TGID,
		event.UID,
		event.GID,
		event.Command,
		event.ContainerRuntime,
		event.ContainerID,
	)

	return nil
}
