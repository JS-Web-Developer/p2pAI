// Package resources: amostra carga do sistema, tempo ocioso e processos de jogo.
package resources

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type Sample struct {
	At          time.Time     `json:"at"`
	CPUPercent  float64       `json:"cpu_percent"`
	MemPercent  float64       `json:"mem_percent"`
	IdleFor     time.Duration `json:"idle_for"`
	GameProcess string        `json:"game_process,omitempty"`
}

type Monitor struct {
	Interval      time.Duration
	GameProcesses []string // nomes em minúsculas; comparação por prefixo

	mu   sync.RWMutex
	last Sample
}

func (m *Monitor) Last() Sample {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.last
}

// Run amostra periodicamente e publica em out (não bloqueia se ninguém lê).
func (m *Monitor) Run(ctx context.Context, out chan<- Sample) {
	if m.Interval == 0 {
		m.Interval = 2 * time.Second
	}
	t := time.NewTicker(m.Interval)
	defer t.Stop()
	for {
		s := m.sample()
		m.mu.Lock()
		m.last = s
		m.mu.Unlock()
		select {
		case out <- s:
		default:
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (m *Monitor) sample() Sample {
	s := Sample{At: time.Now()}
	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		s.CPUPercent = p[0]
	}
	if v, err := mem.VirtualMemory(); err == nil {
		s.MemPercent = v.UsedPercent
	}
	if d, err := idleTime(); err == nil {
		s.IdleFor = d
	} else {
		slog.Debug("idle time indisponível", "err", err)
	}
	s.GameProcess = m.findGame()
	return s
}

func (m *Monitor) findGame() string {
	if len(m.GameProcesses) == 0 {
		return ""
	}
	procs, err := process.Processes()
	if err != nil {
		return ""
	}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		lname := strings.ToLower(name)
		for _, g := range m.GameProcesses {
			if strings.HasPrefix(lname, g) {
				return name
			}
		}
	}
	return ""
}
