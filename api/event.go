package api

import (
	"context"
	"sync"

	"github.com/jessfraz/bpfd/api/grpc"
)

// WARNING:
//
// This buffer does not have a limit, and could potentially grow uncontrollably and put the system in deadlock.
// TODO don't store this in memory

var (
	// events is a variable size buffer.
	events []grpc.Event

	// eventMutex locks the buffer for every transaction.
	eventMutex = sync.Mutex{}
)

// addEvent will add a event to the heap.
func addEvent(event grpc.Event) {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	events = append(events, event)
}

// popEvent will pop a event from the bottom of the heap.
func popEvent() *grpc.Event {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	if len(events) > 0 {
		event := events[len(events)-1]
		events = events[:len(events)-1]
		return &event
	}

	return nil
}

// getEvent will get a event from the bottom of the heap.
func getEvent() *grpc.Event {
	if len(events) > 0 {
		event := events[len(events)-1]
		return &event
	}

	return nil
}

func (s *apiServer) LiveTrace(ctx context.Context, l *grpc.LiveTraceRequest) (*grpc.Event, error) {
	// TODO: make this less shitty.
	// Get an event off the queue (or nil).
	event := getEvent()

	// Handle nil case so we don't get error: "proto: Marshal called with nil".
	if event == nil {
		event = &grpc.Event{}
	}

	// Return it.
	return event, nil
}
