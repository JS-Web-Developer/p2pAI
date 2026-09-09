// Package scheduler: decide se o nó contribui agora.
//
// Regras (docs/PROPOSTA.md, seção 3):
//
//	pausa   CPU > limite por PauseAfter | processo de jogo | janela bloqueada | manual
//	retoma  idle > ResumeIdleAfter e nenhuma regra de pausa ativa
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/p2pai/p2pai/services/daemon/internal/config"
	"github.com/p2pai/p2pai/services/daemon/internal/resources"
)

type State string

const (
	Starting     State = "starting"
	Active       State = "active"
	PausedManual State = "paused_manual"
	PausedAuto   State = "paused_auto"
)

type Status struct {
	State  State     `json:"state"`
	Reason string    `json:"reason,omitempty"`
	Since  time.Time `json:"since"`
}

// Worker é o que o scheduler liga e desliga.
type Worker interface {
	Start(ctx context.Context) error
	Stop()
}

type Scheduler struct {
	Cfg    config.SchedulerConfig
	Worker Worker
	// Enabled: desejo do usuário. false = pausa manual.
	enabled bool

	mu          sync.RWMutex
	status      Status
	highCPUFrom time.Time
	onChange    []func(Status)
}

func New(cfg config.SchedulerConfig, w Worker) *Scheduler {
	return &Scheduler{
		Cfg: cfg, Worker: w, enabled: true,
		status: Status{State: Starting, Since: time.Now()},
	}
}

func (s *Scheduler) OnChange(f func(Status)) { s.onChange = append(s.onChange, f) }

func (s *Scheduler) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// SetEnabled: controle manual. false pausa imediatamente; true volta ao automático.
func (s *Scheduler) SetEnabled(ctx context.Context, v bool) {
	s.mu.Lock()
	s.enabled = v
	s.mu.Unlock()
	if !v {
		s.transition(ctx, PausedManual, "pausado pelo usuário")
	}
}

func (s *Scheduler) Enabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Run consome amostras e aplica as regras.
func (s *Scheduler) Run(ctx context.Context, samples <-chan resources.Sample) {
	for {
		select {
		case <-ctx.Done():
			s.Worker.Stop()
			return
		case smp := <-samples:
			s.evaluate(ctx, smp)
		}
	}
}

func (s *Scheduler) evaluate(ctx context.Context, smp resources.Sample) {
	if !s.Enabled() {
		return
	}
	now := smp.At
	if smp.CPUPercent > s.Cfg.PauseCPUPercent {
		if s.highCPUFrom.IsZero() {
			s.highCPUFrom = now
		}
	} else {
		s.highCPUFrom = time.Time{}
	}

	reason := s.pauseReason(smp, now)
	cur := s.Status()
	switch {
	case reason != "":
		if cur.State != PausedAuto || cur.Reason != reason {
			s.transition(ctx, PausedAuto, reason)
		}
	case cur.State == Active:
		// nada a fazer
	case smp.IdleFor >= s.Cfg.ResumeIdleAfter.D() || cur.State == Starting:
		// Starting: primeira avaliação sem motivo de pausa. Ativa direto para
		// que o usuário veja contribuição logo após instalar.
		s.transition(ctx, Active, "")
	}
}

func (s *Scheduler) pauseReason(smp resources.Sample, now time.Time) string {
	if smp.GameProcess != "" {
		return "jogo em execução: " + smp.GameProcess
	}
	if !s.highCPUFrom.IsZero() && now.Sub(s.highCPUFrom) >= s.Cfg.PauseAfter.D() {
		return "CPU acima do limite"
	}
	for _, r := range s.Cfg.BlockedHours {
		if r.Contains(now) {
			return "janela de horário bloqueada"
		}
	}
	return ""
}

func (s *Scheduler) transition(ctx context.Context, to State, reason string) {
	s.mu.Lock()
	from := s.status
	s.status = Status{State: to, Reason: reason, Since: time.Now()}
	st := s.status
	s.mu.Unlock()

	if to == Active {
		if err := s.Worker.Start(ctx); err != nil {
			slog.Error("iniciar worker", "err", err)
			s.mu.Lock()
			s.status = Status{State: PausedAuto, Reason: "erro: " + err.Error(), Since: time.Now()}
			st = s.status
			s.mu.Unlock()
		}
	} else if from.State == Active || from.State == Starting {
		s.Worker.Stop()
	}
	slog.Info("scheduler", "de", from.State, "para", st.State, "motivo", st.Reason)
	for _, f := range s.onChange {
		f(st)
	}
}
