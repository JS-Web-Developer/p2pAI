// Package api: HTTP/JSON do coordinator.
//
//	POST /v1/nodes/heartbeat   nó registra/atualiza estado (a cada 30 s)
//	GET  /v1/nodes?model=X     nós disponíveis para montar pipeline
//	POST /v1/receipts          cliente envia recibo assinado
//	GET  /v1/leaderboard       top saldos
//	GET  /v1/network           contagem de nós, multiaddrs do relay
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/p2pai/p2pai/services/coordinator/internal/credits"
	"github.com/p2pai/p2pai/services/coordinator/internal/registry"
)

type Server struct {
	Registry   *registry.Registry
	Ledger     *credits.Ledger
	RelayAddrs []string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/nodes/heartbeat", s.heartbeat)
	mux.HandleFunc("GET /v1/nodes", s.nodes)
	mux.HandleFunc("POST /v1/receipts", s.receipt)
	mux.HandleFunc("GET /v1/leaderboard", s.leaderboard)
	mux.HandleFunc("GET /v1/network", s.network)
	return mux
}

func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request) {
	var n registry.Node
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&n); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := peer.Decode(n.PeerID); err != nil {
		http.Error(w, "peer_id inválido", http.StatusBadRequest)
		return
	}
	// TODO(fase 2): exigir assinatura do heartbeat pela chave do nó.
	s.Registry.Upsert(n)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) nodes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Registry.Available(r.URL.Query().Get("model")))
}

func (s *Server) receipt(w http.ResponseWriter, r *http.Request) {
	var rc credits.Receipt
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&rc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if rc.Tokens <= 0 || rc.Nonce == "" || time.Since(rc.IssuedAt) > time.Hour {
		http.Error(w, "recibo inválido", http.StatusBadRequest)
		return
	}
	if !verifyReceipt(rc) {
		http.Error(w, "assinatura inválida", http.StatusUnauthorized)
		return
	}
	if !s.Ledger.Apply(rc) {
		http.Error(w, "recibo repetido", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// verifyReceipt checa a assinatura do cliente sobre o payload canônico.
// Peer IDs Ed25519 embutem a chave pública; não precisa de diretório de chaves.
func verifyReceipt(rc credits.Receipt) bool {
	id, err := peer.Decode(rc.Client)
	if err != nil {
		return false
	}
	pub, err := id.ExtractPublicKey()
	if err != nil || pub == nil {
		return false
	}
	ok, err := pub.Verify(ReceiptPayload(rc), rc.Signature)
	return err == nil && ok
}

// ReceiptPayload é o que o cliente assina. Mesmo formato no daemon.
func ReceiptPayload(rc credits.Receipt) []byte {
	b, _ := json.Marshal(struct {
		Client   string    `json:"client"`
		Node     string    `json:"node"`
		Model    string    `json:"model"`
		Tokens   int64     `json:"tokens"`
		Nonce    string    `json:"nonce"`
		IssuedAt time.Time `json:"issued_at"`
	}{rc.Client, rc.Node, rc.Model, rc.Tokens, rc.Nonce, rc.IssuedAt.UTC().Truncate(time.Second)})
	return b
}


func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Ledger.Top(50))
}

func (s *Server) network(w http.ResponseWriter, r *http.Request) {
	total, active := s.Registry.Count()
	writeJSON(w, map[string]any{
		"nodes_total":  total,
		"nodes_active": active,
		"relays":       s.RelayAddrs,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
