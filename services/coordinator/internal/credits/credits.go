// Package credits: recibos de inferência e saldo por nó. Em memória; Fase 2 persiste.
//
// Recibo: cliente assina (peer_id do cliente, peer_id do nó, tokens, nonce).
// Verificação de assinatura fica no handler da API (precisa da chave pública libp2p).
package credits

import (
	"sync"
	"time"
)

type Receipt struct {
	Client    string    `json:"client"`
	Node      string    `json:"node"`
	Model     string    `json:"model"`
	Tokens    int64     `json:"tokens"`
	Nonce     string    `json:"nonce"`
	IssuedAt  time.Time `json:"issued_at"`
	Signature []byte    `json:"signature"`
}

type Ledger struct {
	mu      sync.Mutex
	balance map[string]int64
	nonces  map[string]struct{}
}

func New() *Ledger {
	return &Ledger{balance: map[string]int64{}, nonces: map[string]struct{}{}}
}

// Apply credita o nó. Rejeita nonce repetido (replay).
func (l *Ledger) Apply(r Receipt) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := r.Client + ":" + r.Nonce
	if _, dup := l.nonces[key]; dup {
		return false
	}
	l.nonces[key] = struct{}{}
	l.balance[r.Node] += r.Tokens
	return true
}

func (l *Ledger) Balance(node string) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balance[node]
}

type Entry struct {
	Node   string `json:"node"`
	Tokens int64  `json:"tokens"`
}

// Top devolve os n maiores saldos (placar).
func (l *Ledger) Top(n int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, 0, len(l.balance))
	for k, v := range l.balance {
		out = append(out, Entry{k, v})
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Tokens > out[j-1].Tokens; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}
