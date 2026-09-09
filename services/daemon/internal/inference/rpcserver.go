// Package inference: processos do llama.cpp (ggml-rpc-server e llama-server).
package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// RPCServer gerencia o ggml-rpc-server local.
type RPCServer struct {
	BinDir   string
	Port     int
	MemoryMB int // 0 = padrão do binário
	Threads  int // 0 = padrão

	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	started time.Time
}

func binName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func (r *RPCServer) Addr() string { return fmt.Sprintf("127.0.0.1:%d", r.Port) }

func (r *RPCServer) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cmd != nil && r.cmd.ProcessState == nil
}

func (r *RPCServer) Uptime() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return 0
	}
	return time.Since(r.started)
}

// Start sobe o processo se não estiver rodando.
func (r *RPCServer) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.ProcessState == nil {
		return nil
	}
	bin := filepath.Join(r.BinDir, binName("ggml-rpc-server"))
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("binário não encontrado: %s", bin)
	}
	args := []string{"-H", "127.0.0.1", "-p", strconv.Itoa(r.Port)}
	if r.MemoryMB > 0 {
		args = append(args, "-m", strconv.Itoa(r.MemoryMB))
	}
	if r.Threads > 0 {
		args = append(args, "-t", strconv.Itoa(r.Threads))
	}
	cctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cctx, bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}
	r.cmd, r.cancel, r.started = cmd, cancel, time.Now()
	slog.Info("ggml-rpc-server iniciado", "pid", cmd.Process.Pid, "port", r.Port)
	go func() {
		err := cmd.Wait()
		r.mu.Lock()
		if r.cmd == cmd {
			r.cmd = nil
		}
		r.mu.Unlock()
		if err != nil && !errors.Is(cctx.Err(), context.Canceled) {
			slog.Warn("ggml-rpc-server encerrou", "err", err)
		}
	}()
	return nil
}

// Stop encerra o processo. Streams em andamento morrem; o cliente refaz o pipeline.
func (r *RPCServer) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return
	}
	r.cancel()
	r.cmd = nil
	slog.Info("ggml-rpc-server parado")
}
