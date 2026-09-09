// Package config: configuração persistente do daemon.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	// Rede
	ListenAddrs []string `json:"listen_addrs"`
	Relays      []string `json:"relays"`      // multiaddrs de relay/bootstrap (coordinator)
	Coordinator string   `json:"coordinator"` // URL HTTP do coordinator (vazio = modo isolado)

	// Inferência
	LlamaBinDir string `json:"llama_bin_dir"` // diretório com ggml-rpc-server e llama-server
	ModelsDir   string `json:"models_dir"`
	RPCPort     int    `json:"rpc_port"`      // porta local do ggml-rpc-server
	RPCMemoryMB int    `json:"rpc_memory_mb"` // 0 = automático

	// Scheduler
	Scheduler SchedulerConfig `json:"scheduler"`

	// API local
	APIAddr string `json:"api_addr"`
}

type SchedulerConfig struct {
	PauseCPUPercent float64     `json:"pause_cpu_percent"` // pausa se CPU acima disso...
	PauseAfter      Duration    `json:"pause_after"`       // ...por este tempo
	ResumeIdleAfter Duration    `json:"resume_idle_after"` // retoma após este tempo sem input
	GameProcesses   []string    `json:"game_processes"`    // nomes (prefixo, case-insensitive)
	BlockedHours    []HourRange `json:"blocked_hours"`     // janelas sem contribuição
	MaxCPUPercent   int         `json:"max_cpu_percent"`   // limite dado ao rpc-server (threads)
}

type HourRange struct {
	From int `json:"from"` // hora inicial, 0-23
	To   int `json:"to"`   // hora final exclusiva, 0-24; From > To atravessa meia-noite
}

func (r HourRange) Contains(t time.Time) bool {
	h := t.Hour()
	if r.From <= r.To {
		return h >= r.From && h < r.To
	}
	return h >= r.From || h < r.To
}

func Default() Config {
	return Config{
		ListenAddrs: []string{"/ip4/0.0.0.0/udp/0/quic-v1", "/ip4/0.0.0.0/tcp/0"},
		RPCPort:     50052,
		APIAddr:     "127.0.0.1:7433",
		Scheduler: SchedulerConfig{
			PauseCPUPercent: 80,
			PauseAfter:      Duration(10 * time.Second),
			ResumeIdleAfter: Duration(5 * time.Minute),
			GameProcesses: []string{
				"steam", "steamwebhelper", "epicgameslauncher", "battle.net",
				"riotclientservices", "league of legends", "valorant", "cs2", "dota2",
			},
			MaxCPUPercent: 50,
		},
	}
}

// Dir devolve o diretório de dados do usuário para o P2PAI.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(base, "p2pai")
	return d, os.MkdirAll(d, 0o700)
}

func Load(path string) (Config, error) {
	cfg := Default()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, cfg.Save(path)
	}
	if err != nil {
		return cfg, err
	}
	return cfg, json.Unmarshal(b, &cfg)
}

func (c Config) Save(path string) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Duration serializa como string legível ("10s", "5m") no JSON.
type Duration time.Duration

func (d Duration) D() time.Duration { return time.Duration(d) }

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		var n int64
		if err2 := json.Unmarshal(b, &n); err2 != nil {
			return err
		}
		*d = Duration(n)
		return nil
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}
