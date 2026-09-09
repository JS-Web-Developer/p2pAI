// p2pai-daemon: nó da rede P2PAI.
//
// Liga: config -> host libp2p -> monitor de recursos -> scheduler -> API local.
// O scheduler controla o ggml-rpc-server; o host encaminha streams a ele.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/p2pai/p2pai/services/daemon/internal/api"
	"github.com/p2pai/p2pai/services/daemon/internal/config"
	"github.com/p2pai/p2pai/services/daemon/internal/inference"
	"github.com/p2pai/p2pai/services/daemon/internal/p2p"
	"github.com/p2pai/p2pai/services/daemon/internal/resources"
	"github.com/p2pai/p2pai/services/daemon/internal/scheduler"
	"github.com/p2pai/p2pai/services/daemon/internal/stats"
)

var version = "dev"

// worker adapta RPCServer + Node ao contrato do scheduler: ao ativar, sobe o
// rpc-server e passa a aceitar streams; ao pausar, recusa streams e derruba o processo.
type worker struct {
	rpc  *inference.RPCServer
	node *p2p.Node
}

func (w *worker) Start(ctx context.Context) error {
	if err := w.rpc.Start(ctx); err != nil {
		return err
	}
	w.node.SetRPCTarget(w.rpc.Addr())
	w.node.SetAccepting(true)
	return nil
}

func (w *worker) Stop() {
	w.node.SetAccepting(false)
	w.rpc.Stop()
}

func main() {
	dir, err := config.Dir()
	if err != nil {
		fatal(err)
	}
	cfgPath := flag.String("config", filepath.Join(dir, "config.json"), "arquivo de configuração")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fatal(err)
	}
	if cfg.LlamaBinDir == "" {
		// Padrão: binários ao lado do daemon (empacotamento).
		exe, _ := os.Executable()
		cfg.LlamaBinDir = filepath.Dir(exe)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	node, err := p2p.New(ctx, p2p.Options{
		KeyFile:     filepath.Join(dir, "node.key"),
		ListenAddrs: cfg.ListenAddrs,
		Relays:      cfg.Relays,
	})
	if err != nil {
		fatal(err)
	}
	defer node.Close()
	slog.Info("nó", "peer_id", node.Host.ID())

	threads := runtime.NumCPU() * cfg.Scheduler.MaxCPUPercent / 100
	if threads < 1 {
		threads = 1
	}
	rpc := &inference.RPCServer{
		BinDir: cfg.LlamaBinDir, Port: cfg.RPCPort, MemoryMB: cfg.RPCMemoryMB, Threads: threads,
	}

	mon := &resources.Monitor{Interval: 2 * time.Second, GameProcesses: cfg.Scheduler.GameProcesses}
	samples := make(chan resources.Sample, 1)
	go mon.Run(ctx, samples)

	sched := scheduler.New(cfg.Scheduler, &worker{rpc: rpc, node: node})
	go sched.Run(ctx, samples)

	store, err := stats.Open(filepath.Join(dir, "stats.json"))
	if err != nil {
		fatal(err)
	}
	go telemetry(ctx, sched, node, store)

	srv := &http.Server{
		Addr: cfg.APIAddr,
		Handler: (&api.Server{
			Sched: sched, Node: node, Monitor: mon, Stats: store, Version: version,
		}).Handler(),
	}
	go func() {
		<-ctx.Done()
		sctx, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()
		srv.Shutdown(sctx)
	}()
	slog.Info("API local", "addr", cfg.APIAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatal(err)
	}
}

// telemetry acumula segundos ativos e tráfego a cada 10 s.
func telemetry(ctx context.Context, s *scheduler.Scheduler, n *p2p.Node, st *stats.Store) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	var lastNet p2p.Stats
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cur := n.Stats()
			active := s.Status().State == scheduler.Active
			st.Update(func(x *stats.Stats) {
				if active {
					x.ActiveSeconds += 10
				}
				x.RequestsServed += cur.StreamsTotal - lastNet.StreamsTotal
				x.BytesIn += cur.BytesIn - lastNet.BytesIn
				x.BytesOut += cur.BytesOut - lastNet.BytesOut
			})
			lastNet = cur
		}
	}
}

func fatal(err error) {
	slog.Error("fatal", "err", err)
	os.Exit(1)
}
