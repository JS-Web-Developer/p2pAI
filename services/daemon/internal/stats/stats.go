// Package stats: contadores de contribuição persistidos em disco.
package stats

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Stats struct {
	ActiveSeconds  int64     `json:"active_seconds"`
	RequestsServed int64     `json:"requests_served"`
	BytesIn        int64     `json:"bytes_in"`
	BytesOut       int64     `json:"bytes_out"`
	FirstSeen      time.Time `json:"first_seen"`
	LastUpdated    time.Time `json:"last_updated"`
}

type Store struct {
	path string
	mu   sync.Mutex
	s    Stats
}

func Open(path string) (*Store, error) {
	st := &Store{path: path}
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, &st.s); err != nil {
			return nil, err
		}
	}
	if st.s.FirstSeen.IsZero() {
		st.s.FirstSeen = time.Now()
	}
	return st, nil
}

func (st *Store) Get() Stats {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.s
}

// Update aplica f e grava. Chamado pelo loop de telemetria a cada poucos segundos.
func (st *Store) Update(f func(*Stats)) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	f(&st.s)
	st.s.LastUpdated = time.Now()
	b, err := json.MarshalIndent(st.s, "", "  ")
	if err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}
