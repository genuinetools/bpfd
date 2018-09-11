package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jessfraz/bpfd/action"
	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/program"
	rulespkg "github.com/jessfraz/bpfd/rules"
	"github.com/sirupsen/logrus"
)

var (
	// TODO: don't store these in memory
	rules map[string]map[string]grpc.Rule
	// rulesMutex locks the buffer for every transaction.
	rulesMutex = sync.Mutex{}
)

type apiServer struct{}

// NewServer returns grpc server instance.
func NewServer(r map[string]map[string]grpc.Rule) (grpc.APIServer, error) {
	rules = r

	// List all the compiled in programs.
	programs := program.List()
	logrus.Infof("Daemon compiled with programs: %s", strings.Join(programs, ", "))

	// List all the compiled in actions.
	actions := action.List()
	logrus.Infof("Daemon compiled with actions: %s", strings.Join(actions, ", "))

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

				if event == nil || event.Data == nil {
					continue
				}

				event.ContainerRuntime = string(proc.GetContainerRuntime(int(event.TGID), int(event.PID)))

				progRules, _ := rules[p]

				if len(progRules) < 1 {
					// Just send to stdout and be done with it.
					action, err := action.Get("stdout")
					if err != nil {
						logrus.Warn(err)
						continue
					}
					action.Do(event)
					continue
				}

				for _, rule := range progRules {
					// Verify the event matches for the rules.
					match, actions := rulespkg.Match(rule, event.Data, event.ContainerRuntime)
					if !match {
						// We didn't find what we were searching for so continue.
						continue
					}

					event.ContainerID = proc.GetContainerID(int(event.TGID), int(event.PID))
					event.Program = p

					// Add this event to our queue of events.
					// addEvent(*event)

					// Perform the actions.
					for _, a := range actions {
						action, err := action.Get(a)
						if err != nil {
							logrus.Warn(err)
							continue
						}
						action.Do(event)
					}

					// Remove the event from the queue.
					// popEvent()
				}

			}
		}(p, prog)

		// Start the program.
		prog.Start()
		logrus.Infof("Watching events for plugin %s", p)
	}

	return &apiServer{}, nil
}

func (s *apiServer) CreateRule(ctx context.Context, c *grpc.CreateRuleRequest) (*grpc.CreateRuleResponse, error) {
	rulesMutex.Lock()
	defer rulesMutex.Unlock()

	if c == nil || c.Rule == nil {
		return nil, errors.New("rule cannot be nil")
	}

	if len(c.Rule.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	if len(c.Rule.Program) < 1 {
		return nil, errors.New("rule program cannot be empty")
	}

	logrus.WithFields(logrus.Fields{
		"program": c.Rule.Program,
		"name":    c.Rule.Name,
	}).Infof("Created rule")

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
	rulesMutex.Lock()
	defer rulesMutex.Unlock()

	if r == nil || len(r.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	// If they passed the program then only remove the rule for that program.
	if len(r.Program) > 0 {
		delete(rules[r.Program], r.Name)

		logrus.WithFields(logrus.Fields{
			"program": r.Program,
			"name":    r.Name,
		}).Infof("Deleted rule")

		return &grpc.RemoveRuleResponse{}, nil
	}

	// Iterate over the programs and find the rule.
	for p, prs := range rules {
		for name := range prs {
			if name == r.Name {
				delete(rules[p], r.Name)

				logrus.WithFields(logrus.Fields{
					"program": p,
					"name":    r.Name,
				}).Infof("Deleted rule")

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
