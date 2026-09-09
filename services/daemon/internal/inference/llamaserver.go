package inference

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LlamaServer sobe um llama-server (API OpenAI-compatível) apontando para
// rpc-servers remotos expostos localmente pela camada p2p. Lado cliente.
type LlamaServer struct {
	BinDir    string
	ModelPath string
	Port      int
	RPCAddrs  []string // endereços locais expostos por p2p.Expose

	mu     sync.Mutex
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func (l *LlamaServer) URL() string { return fmt.Sprintf("http://127.0.0.1:%d", l.Port) }

func (l *LlamaServer) Start(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cmd != nil && l.cmd.ProcessState == nil {
		return nil
	}
	bin := filepath.Join(l.BinDir, binName("llama-server"))
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("binário não encontrado: %s", bin)
	}
	if _, err := os.Stat(l.ModelPath); err != nil {
		return fmt.Errorf("modelo não encontrado: %s", l.ModelPath)
	}
	args := []string{
		"-m", l.ModelPath,
		"--host", "127.0.0.1", "--port", strconv.Itoa(l.Port),
		"-ngl", "99",
	}
	if len(l.RPCAddrs) > 0 {
		args = append(args, "--rpc", strings.Join(l.RPCAddrs, ","))
	}
	cctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cctx, bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}
	l.cmd, l.cancel = cmd, cancel
	go func() {
		cmd.Wait()
		l.mu.Lock()
		if l.cmd == cmd {
			l.cmd = nil
		}
		l.mu.Unlock()
	}()
	return l.waitReady(ctx, 120*time.Second)
}

func (l *LlamaServer) waitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", l.Port), time.Second)
		if err == nil {
			c.Close()
			resp, err := http.Get(l.URL() + "/health")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("llama-server não ficou pronto em %s", timeout)
}

func (l *LlamaServer) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cmd == nil {
		return
	}
	l.cancel()
	l.cmd = nil
}
