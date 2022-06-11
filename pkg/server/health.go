package server

import (
	"fmt"
	"net/http"
)

// GetHealth get health of the server
func (s *Server) GetHealth() bool {
	s.m.Lock()
	defer s.m.Unlock()
	for _, h := range s.health {
		if !h {
			return false
		}
	}
	return true
}

// SetHealth set health of a component
func (s *Server) SetHealth(key string, health bool) {
	s.m.Lock()
	defer s.m.Unlock()
	s.health[key] = health
}

// HealthHandler handles health requests
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if !s.GetHealth() {
		http.Error(w, "not ok", http.StatusInternalServerError)

		return
	}
	fmt.Fprintf(w, "ok")
}
