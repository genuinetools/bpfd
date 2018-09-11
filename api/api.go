package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/program"
	"github.com/sirupsen/logrus"
)

var (
	// TODO: maybe don't store these in memory
	rules map[string]map[string]grpc.Rule
)

type apiServer struct{}

// NewServer returns grpc server instance.
func NewServer(r map[string]map[string]grpc.Rule) (grpc.APIServer, error) {
	rules = r

	// List all the compiled in programs.
	programs := program.List()
	logrus.Infof("Daemon compiled with programs: %s", strings.Join(programs, ", "))

	// Load all the compiled in programs.
	for _, p := range programs {
		// We can ignore the error below since we are using the list from our code
		// so the program has to exist in the map.
		prog, _ := program.Get(p)
		if err := prog.Load(); err != nil {
			return nil, fmt.Errorf("loading program %s failed: %v", p, err)
		}

		go func(p string, prog program.Program) {
			for {
				// Watch the events for the program.
				event, err := prog.WatchEvent()
				if err != nil {
					logrus.Warnf("watch event for program %s failed: %v", p, err)
				}

				if event == nil {
					continue
				}

				runtime := proc.GetContainerRuntime(int(event.TGID), int(event.PID))

				progRules, _ := rules[p]

				// Verify the event matches for the rules.
				if !program.Match(progRules, event.Data, runtime) {
					// We didn't find what we were searching for so continue.
					continue
				}

				logrus.WithFields(logrus.Fields{
					"program":           p,
					"pid":               fmt.Sprintf("%d", event.PID),
					"tgid":              fmt.Sprintf("%d", event.TGID),
					"container_runtime": string(runtime),
					"container_id":      proc.GetContainerID(int(event.TGID), int(event.PID)),
				}).Infof("%#v", event.Data)
			}
		}(p, prog)

		// Start the program.
		prog.Start()
		logrus.Infof("Watching events for plugin %s", p)
	}

	return &apiServer{}, nil
}

func (s *apiServer) CreateRule(ctx context.Context, c *grpc.CreateRuleRequest) (*grpc.CreateRuleResponse, error) {
	if c == nil || c.Rule == nil {
		return nil, errors.New("rule cannot be nil")
	}

	if len(c.Rule.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	if len(c.Rule.Program) < 1 {
		return nil, errors.New("rule program cannot be empty")
	}

	// Check if we already have rules for the program to avoid a panic.
	_, ok := rules[c.Rule.Program]
	if !ok {
		rules[c.Rule.Program] = map[string]grpc.Rule{c.Rule.Name: *c.Rule}
		return &grpc.CreateRuleResponse{}, nil
	}

	// Add the rule to our existing rules for the program.
	// TODO: decide to error or not on overwrite
	rules[c.Rule.Program][c.Rule.Name] = *c.Rule
	return &grpc.CreateRuleResponse{}, nil
}

// TODO: find a better way to remove without program
func (s *apiServer) RemoveRule(ctx context.Context, r *grpc.RemoveRuleRequest) (*grpc.RemoveRuleResponse, error) {
	if r == nil || len(r.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	// If they passed the program then only remove the rule for that program.
	if len(r.Program) > 0 {
		delete(rules[r.Program], r.Name)
		return &grpc.RemoveRuleResponse{}, nil
	}

	// Iterate over the programs and find the rule.
	for p, prs := range rules {
		for name := range prs {
			if name == r.Name {
				delete(rules[p], r.Name)
				continue
			}
		}
	}
	return &grpc.RemoveRuleResponse{}, nil
}

func (s *apiServer) ListRules(ctx context.Context, r *grpc.ListRulesRequest) (*grpc.ListRulesResponse, error) {
	var rs []*grpc.Rule

	for _, prs := range rules {
		for _, rule := range prs {
			rs = append(rs, &rule)
		}
	}

	return &grpc.ListRulesResponse{Rules: rs}, nil
}
