package api

import (
	"github.com/genuinetools/bpfd/api/grpc"
)

// addEvent will add a event to the heap.
func (s *apiServer) addEvent(event grpc.Event) {
	if !s.isStreaming {
		// Return early if we are not streaming.
		return
	}

	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	s.events = append(s.events, event)
}

// popEvent will pop a event from the bottom of the heap.
func (s *apiServer) popEvent() *grpc.Event {
	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	if len(s.events) > 0 {
		event := s.events[len(s.events)-1]
		s.events = s.events[:len(s.events)-1]
		return &event
	}

	return nil
}
