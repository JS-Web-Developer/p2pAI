// p2pai-coordinator: bootstrap libp2p + relay (Circuit Relay v2) + API HTTP.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"

	"github.com/p2pai/p2pai/services/coordinator/internal/api"
	"github.com/p2pai/p2pai/services/coordinator/internal/credits"
	"github.com/p2pai/p2pai/services/coordinator/internal/registry"
)

func main() {
	apiAddr := flag.String("api", ":8080", "endereço HTTP")
	keyFile := flag.String("key", "coordinator.key", "chave privada libp2p")
	quicPort := flag.Int("quic", 4001, "porta QUIC")
	tcpPort := flag.Int("tcp", 4001, "porta TCP")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	priv, err := loadOrCreateKey(*keyFile)
	if err != nil {
		fatal(err)
	}
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", *quicPort),
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *tcpPort),
		),
		libp2p.EnableRelayService(),
		libp2p.ForceReachabilityPublic(),
	)
	if err != nil {
		fatal(err)
	}
	defer h.Close()
	if _, err := relay.New(h, relay.WithInfiniteLimits()); err != nil {
		fatal(err)
	}
	var addrs []string
	for _, a := range h.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", a, h.ID()))
	}
	slog.Info("relay", "peer_id", h.ID(), "addrs", addrs)

	reg := registry.New(90 * time.Second)
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				reg.Sweep()
			}
		}
	}()

	srv := &http.Server{
		Addr:              *apiAddr,
		Handler:           (&api.Server{Registry: reg, Ledger: credits.New(), RelayAddrs: addrs}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		sctx, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()
		srv.Shutdown(sctx)
	}()
	slog.Info("API", "addr", *apiAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatal(err)
	}
}

func loadOrCreateKey(path string) (crypto.PrivKey, error) {
	if b, err := os.ReadFile(path); err == nil {
		return crypto.UnmarshalPrivateKey(b)
	}
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	b, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return priv, os.WriteFile(path, b, 0o600)
}

func fatal(err error) {
	slog.Error("fatal", "err", err)
	os.Exit(1)
}
