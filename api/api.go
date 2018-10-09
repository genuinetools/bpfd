package api

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/genuinetools/bpfd/action"
	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/genuinetools/bpfd/proc"
	rulespkg "github.com/genuinetools/bpfd/rules"
	"github.com/genuinetools/bpfd/tracer"
	"github.com/sirupsen/logrus"
)

type apiServer struct {
	// TODO: don't store these in memory
	rules map[string]map[string]grpc.Rule
	// rulesMutex locks the buffer for every transaction.
	rulesMutex sync.Mutex

	tracers map[string]tracer.Tracer
	actions map[string]action.Action

	tracerList []string
	actionList []string

	isStreaming bool

	// events is a variable size buffer.
	events []grpc.Event
	// eventMutex locks the buffer for every transaction.
	eventsMutex sync.Mutex
}

// Opts holds the options for a server.
type Opts struct {
	Rules   map[string]map[string]grpc.Rule
	Tracers map[string]tracer.Tracer
	Actions map[string]action.Action

	TracerList []string
	ActionList []string
}

// NewServer returns grpc server instance.
func NewServer(ctx context.Context, opt Opts) (grpc.APIServer, error) {
	server := &apiServer{
		rules:      opt.Rules,
		tracers:    opt.Tracers,
		actions:    opt.Actions,
		tracerList: opt.TracerList,
		actionList: opt.ActionList,
	}

	// Load all the compiled in tracers.
	for p, prog := range opt.Tracers {
		if err := prog.Load(); err != nil {
			return nil, fmt.Errorf("loading tracer %s failed: %v", p, err)
		}

		go func(p string, prog tracer.Tracer) {
			for {
				// Watch the events for the tracer.
				event, err := prog.WatchEvent(ctx)
				if err != nil {
					logrus.Warnf("watch event for tracer %s failed: %v", p, err)
				}

				if event == nil || event.Data == nil {
					continue
				}

				// Get the container runtime if we don't already have it.
				if len(event.ContainerRuntime) < 1 {
					event.ContainerRuntime = string(proc.GetContainerRuntime(int(event.TGID), int(event.PID)))
				}
				// Get the container ID if we don't already have it.
				if len(event.ContainerID) < 1 {
					event.ContainerID = proc.GetContainerID(int(event.TGID), int(event.PID))
				}
				event.Tracer = p

				progRules, _ := server.rules[p]

				if len(progRules) < 1 {
					// Just send to stdout and be done with it.
					opt.Actions["stdout"].Do(event)

					// Add this event to our queue of events if we are streaming.
					server.addEvent(*event)
					continue
				}

				for _, rule := range progRules {
					// Verify the event matches for the rules.
					if match := rulespkg.Match(rule, event.Data, event.ContainerRuntime); !match {
						// We didn't find what we were searching for so continue.
						continue
					}

					// Add this event to our queue of events if we are streaming.
					server.addEvent(*event)

					// Perform the actions.
					for _, a := range rule.Actions {
						action, ok := opt.Actions[a]
						if !ok {
							logrus.Warnf("action %s provided by rule %s does not exist", a, rule.Name)
							continue
						}
						action.Do(event)
					}
				}

			}
		}(p, prog)

		// Start the tracer.
		prog.Start()
		logrus.Infof("Watching events for plugin: %s", p)
	}

	return server, nil
}

func (s *apiServer) CreateRule(ctx context.Context, c *grpc.CreateRuleRequest) (*grpc.CreateRuleResponse, error) {
	s.rulesMutex.Lock()
	defer s.rulesMutex.Unlock()

	if c == nil || c.Rule == nil {
		return nil, errors.New("rule cannot be nil")
	}

	// Validate the rule.
	if err := rulespkg.Validate(*c.Rule); err != nil {
		return nil, err
	}
	if err := rulespkg.ValidateTracersAndActions(*c.Rule, s.tracerList, s.actionList); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"tracer": c.Rule.Tracer,
		"name":   c.Rule.Name,
	}).Infof("Created rule")

	// Check if we already have rules for the tracer to avoid a panic.
	_, ok := s.rules[c.Rule.Tracer]
	if !ok {
		s.rules[c.Rule.Tracer] = map[string]grpc.Rule{c.Rule.Name: *c.Rule}
		return &grpc.CreateRuleResponse{}, nil
	}

	// Add the rule to our existing rules for the tracer.
	// TODO: decide to error or not on overwrite
	s.rules[c.Rule.Tracer][c.Rule.Name] = *c.Rule
	return &grpc.CreateRuleResponse{}, nil
}

// TODO: find a better way to remove without tracer
func (s *apiServer) RemoveRule(ctx context.Context, r *grpc.RemoveRuleRequest) (*grpc.RemoveRuleResponse, error) {
	s.rulesMutex.Lock()
	defer s.rulesMutex.Unlock()

	if r == nil || len(r.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	// If they passed the tracer then only remove the rule for that tracer.
	if len(r.Tracer) > 0 {
		delete(s.rules[r.Tracer], r.Name)

		logrus.WithFields(logrus.Fields{
			"tracer": r.Tracer,
			"name":   r.Name,
		}).Infof("Deleted rule")

		return &grpc.RemoveRuleResponse{}, nil
	}

	// Iterate over the tracers and find the rule.
	for p, prs := range s.rules {
		for name := range prs {
			if name == r.Name {
				delete(s.rules[p], r.Name)

				logrus.WithFields(logrus.Fields{
					"tracer": p,
					"name":   r.Name,
				}).Infof("Deleted rule")

				continue
			}
		}
	}
	return &grpc.RemoveRuleResponse{}, nil
}

func (s *apiServer) ListRules(ctx context.Context, r *grpc.ListRulesRequest) (*grpc.ListRulesResponse, error) {
	var rs []*grpc.Rule

	for _, prs := range s.rules {
		for _, rule := range prs {
			rs = append(rs, &rule)
		}
	}

	return &grpc.ListRulesResponse{Rules: rs}, nil
}

func (s *apiServer) LiveTrace(l *grpc.LiveTraceRequest, stream grpc.API_LiveTraceServer) error {
	logrus.Debug("live trace streaming client CONNECTED")
	s.isStreaming = true

	for s.isStreaming {
		if err := stream.Context().Err(); err != nil {
			s.isStreaming = false
			break
		}

		event := s.popEvent()
		if event == nil {
			continue
		}

		if err := stream.Send(event); err != nil {
			s.isStreaming = false
			logrus.Debug("live trace streaming client DISCONNECTED")
			return err
		}

	}

	logrus.Debug("live trace streaming client DISCONNECTED")
	return nil
}
