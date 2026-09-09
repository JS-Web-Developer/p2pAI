// Package api: HTTP/JSON local para a UI (Electron).
//
//	GET  /v1/status        estado do scheduler, rede, dispositivo
//	POST /v1/contribution  {"enabled": bool}
//	GET  /v1/stats         contadores
//	GET  /v1/resources     última amostra do monitor
//	GET  /v1/events        SSE: mudanças de estado
package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"

	"github.com/p2pai/p2pai/services/daemon/internal/p2p"
	"github.com/p2pai/p2pai/services/daemon/internal/resources"
	"github.com/p2pai/p2pai/services/daemon/internal/scheduler"
	"github.com/p2pai/p2pai/services/daemon/internal/stats"
)

type Server struct {
	Sched   *scheduler.Scheduler
	Node    *p2p.Node
	Monitor *resources.Monitor
	Stats   *stats.Store
	Version string

	mu   sync.Mutex
	subs map[chan scheduler.Status]struct{}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/status", s.status)
	mux.HandleFunc("POST /v1/contribution", s.contribution)
	mux.HandleFunc("GET /v1/stats", s.stats)
	mux.HandleFunc("GET /v1/resources", s.resources)
	mux.HandleFunc("GET /v1/events", s.events)
	s.subs = map[chan scheduler.Status]struct{}{}
	s.Sched.OnChange(s.broadcast)
	return localOnly(mux)
}

// localOnly rejeita qualquer origem que não seja loopback. Defesa extra além do bind.
func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" && !isLocalOrigin(o) {
			http.Error(w, "origem não permitida", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalOrigin(o string) bool {
	for _, p := range []string{"http://localhost", "http://127.0.0.1", "file://", "app://"} {
		if len(o) >= len(p) && o[:len(p)] == p {
			return true
		}
	}
	return false
}

type statusResponse struct {
	Scheduler scheduler.Status `json:"scheduler"`
	Network   p2p.Stats        `json:"network"`
	Device    deviceInfo       `json:"device"`
	Version   string           `json:"version"`
}

type deviceInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	CPUs int    `json:"cpus"`
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, statusResponse{
		Scheduler: s.Sched.Status(),
		Network:   s.Node.Stats(),
		Device:    deviceInfo{OS: runtime.GOOS, Arch: runtime.GOARCH, CPUs: runtime.NumCPU()},
		Version:   s.Version,
	})
}

func (s *Server) contribution(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.Sched.SetEnabled(r.Context(), body.Enabled)
	writeJSON(w, s.Sched.Status())
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Stats.Get())
}

func (s *Server) resources(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Monitor.Last())
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming não suportado", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan scheduler.Status, 8)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()
	send := func(st scheduler.Status) {
		b, _ := json.Marshal(st)
		w.Write([]byte("event: scheduler\ndata: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		fl.Flush()
	}
	send(s.Sched.Status())
	for {
		select {
		case <-r.Context().Done():
			return
		case st := <-ch:
			send(st)
		}
	}
}

func (s *Server) broadcast(st scheduler.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- st:
		default:
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
