// Package registry: nós conhecidos pelo coordinator, em memória.
// Fase 2 troca por Postgres.
package registry

import (
	"sort"
	"sync"
	"time"
)

type Node struct {
	PeerID    string    `json:"peer_id"`
	Addrs     []string  `json:"addrs"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Tier      int       `json:"tier"`
	RAMBytes  int64     `json:"ram_bytes"`
	VRAMBytes int64     `json:"vram_bytes"`
	Models    []string  `json:"models"` // modelos que o nó tem em disco
	Active    bool      `json:"active"` // scheduler em estado active
	LastSeen  time.Time `json:"last_seen"`
	FirstSeen time.Time `json:"first_seen"`
}

type Registry struct {
	TTL time.Duration

	mu    sync.RWMutex
	nodes map[string]*Node
}

func New(ttl time.Duration) *Registry {
	return &Registry{TTL: ttl, nodes: map[string]*Node{}}
}

// Upsert registra ou atualiza (heartbeat).
func (r *Registry) Upsert(n Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if cur, ok := r.nodes[n.PeerID]; ok {
		n.FirstSeen = cur.FirstSeen
	} else {
		n.FirstSeen = now
	}
	n.LastSeen = now
	r.nodes[n.PeerID] = &n
}

// Available lista nós ativos, vistos dentro do TTL, ordenados por tier (melhor primeiro).
func (r *Registry) Available(model string) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cut := time.Now().Add(-r.TTL)
	var out []Node
	for _, n := range r.nodes {
		if !n.Active || n.LastSeen.Before(cut) {
			continue
		}
		if model != "" && !contains(n.Models, model) {
			continue
		}
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier < out[j].Tier
		}
		return out[i].PeerID < out[j].PeerID
	})
	return out
}

func (r *Registry) Count() (total, active int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cut := time.Now().Add(-r.TTL)
	for _, n := range r.nodes {
		if n.LastSeen.Before(cut) {
			continue
		}
		total++
		if n.Active {
			active++
		}
	}
	return
}

// Sweep remove nós expirados há mais de 10x TTL.
func (r *Registry) Sweep() {
	r.mu.Lock()
	defer r.mu.Unlock()
	cut := time.Now().Add(-10 * r.TTL)
	for id, n := range r.nodes {
		if n.LastSeen.Before(cut) {
			delete(r.nodes, id)
		}
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
