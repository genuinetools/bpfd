package api

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jessfraz/bpfd/action"
	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/program"
	rulespkg "github.com/jessfraz/bpfd/rules"
	"github.com/sirupsen/logrus"
)

type apiServer struct {
	// TODO: don't store these in memory
	rules map[string]map[string]grpc.Rule
	// rulesMutex locks the buffer for every transaction.
	rulesMutex sync.Mutex

	programs map[string]program.Program
	actions  map[string]action.Action

	programList []string
	actionList  []string
}

// Opts holds the options for a server.
type Opts struct {
	Rules    map[string]map[string]grpc.Rule
	Programs map[string]program.Program
	Actions  map[string]action.Action

	ProgramList []string
	ActionList  []string
}

// NewServer returns grpc server instance.
func NewServer(opt Opts) (grpc.APIServer, error) {
	server := &apiServer{
		rules:       opt.Rules,
		programs:    opt.Programs,
		actions:     opt.Actions,
		programList: opt.ProgramList,
		actionList:  opt.ActionList,
	}

	// Load all the compiled in programs.
	for p, prog := range opt.Programs {
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

				progRules, _ := server.rules[p]

				if len(progRules) < 1 {
					// Just send to stdout and be done with it.
					opt.Actions["stdout"].Do(event)
					continue
				}

				for _, rule := range progRules {
					// Verify the event matches for the rules.
					if match := rulespkg.Match(rule, event.Data, event.ContainerRuntime); !match {
						// We didn't find what we were searching for so continue.
						continue
					}

					event.ContainerID = proc.GetContainerID(int(event.TGID), int(event.PID))
					event.Program = p

					// Add this event to our queue of events.
					// addEvent(*event)

					// Perform the actions.
					for _, a := range rule.Actions {
						action, ok := opt.Actions[a]
						if !ok {
							logrus.Warnf("action %s provided by rule %s does not exist", a, rule.Name)
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
	if err := rulespkg.ValidateProgramsAndActions(*c.Rule, s.programList, s.actionList); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"program": c.Rule.Program,
		"name":    c.Rule.Name,
	}).Infof("Created rule")

	// Check if we already have rules for the program to avoid a panic.
	_, ok := s.rules[c.Rule.Program]
	if !ok {
		s.rules[c.Rule.Program] = map[string]grpc.Rule{c.Rule.Name: *c.Rule}
		return &grpc.CreateRuleResponse{}, nil
	}

	// Add the rule to our existing rules for the program.
	// TODO: decide to error or not on overwrite
	s.rules[c.Rule.Program][c.Rule.Name] = *c.Rule
	return &grpc.CreateRuleResponse{}, nil
}

// TODO: find a better way to remove without program
func (s *apiServer) RemoveRule(ctx context.Context, r *grpc.RemoveRuleRequest) (*grpc.RemoveRuleResponse, error) {
	s.rulesMutex.Lock()
	defer s.rulesMutex.Unlock()

	if r == nil || len(r.Name) < 1 {
		return nil, errors.New("rule name cannot be empty")
	}

	// If they passed the program then only remove the rule for that program.
	if len(r.Program) > 0 {
		delete(s.rules[r.Program], r.Name)

		logrus.WithFields(logrus.Fields{
			"program": r.Program,
			"name":    r.Name,
		}).Infof("Deleted rule")

		return &grpc.RemoveRuleResponse{}, nil
	}

	// Iterate over the programs and find the rule.
	for p, prs := range s.rules {
		for name := range prs {
			if name == r.Name {
				delete(s.rules[p], r.Name)

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

	for _, prs := range s.rules {
		for _, rule := range prs {
			rs = append(rs, &rule)
		}
	}

	return &grpc.ListRulesResponse{Rules: rs}, nil
}
