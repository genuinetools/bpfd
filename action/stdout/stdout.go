package stdout

import (
	"fmt"

	"github.com/jessfraz/bpfd/action"
	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"tracer":            event.Tracer,
		"pid":               fmt.Sprintf("%d", event.PID),
		"tgid":              fmt.Sprintf("%d", event.TGID),
		"container_runtime": string(event.ContainerRuntime),
		"container_id":      event.ContainerID,
	}).Infof("%#v", event.Data)

	return nil
}
