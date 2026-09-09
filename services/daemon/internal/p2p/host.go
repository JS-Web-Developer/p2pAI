// Package p2p: host libp2p do nó.
//
// Protocolos:
//
//	/p2pai/rpc/0.1.0  stream bruto encaminhado ao ggml-rpc-server local
package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const RPCProtocol protocol.ID = "/p2pai/rpc/0.1.0"

type Options struct {
	KeyFile     string
	ListenAddrs []string
	Relays      []string
}

// Node encapsula o host e o encaminhamento de streams.
type Node struct {
	Host host.Host

	mu        sync.RWMutex
	rpcTarget string // endereço TCP local do rpc-server; vazio = recusar streams
	accepting atomic.Bool

	streamsOpen  atomic.Int64
	streamsTotal atomic.Int64
	bytesIn      atomic.Int64
	bytesOut     atomic.Int64
}

func New(ctx context.Context, o Options) (*Node, error) {
	priv, err := loadOrCreateKey(o.KeyFile)
	if err != nil {
		return nil, err
	}
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(o.ListenAddrs...),
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
	}
	var relays []peer.AddrInfo
	for _, r := range o.Relays {
		ai, err := peer.AddrInfoFromString(r)
		if err != nil {
			return nil, fmt.Errorf("relay %q: %w", r, err)
		}
		relays = append(relays, *ai)
	}
	if len(relays) > 0 {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(relays))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	n := &Node{Host: h}
	h.SetStreamHandler(RPCProtocol, n.handleRPC)

	for _, ai := range relays {
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		if err := h.Connect(cctx, ai); err != nil {
			slog.Warn("relay indisponível", "peer", ai.ID, "err", err)
		}
		cancel()
	}
	return n, nil
}

func (n *Node) Close() error { return n.Host.Close() }

// SetRPCTarget define para onde streams recebidos são encaminhados.
func (n *Node) SetRPCTarget(addr string) {
	n.mu.Lock()
	n.rpcTarget = addr
	n.mu.Unlock()
}

// SetAccepting liga/desliga aceitação de trabalho (scheduler).
func (n *Node) SetAccepting(v bool) { n.accepting.Store(v) }

func (n *Node) handleRPC(s network.Stream) {
	n.mu.RLock()
	target := n.rpcTarget
	n.mu.RUnlock()
	if !n.accepting.Load() || target == "" {
		s.Reset()
		return
	}
	c, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		slog.Error("rpc-server local indisponível", "err", err)
		s.Reset()
		return
	}
	n.streamsOpen.Add(1)
	n.streamsTotal.Add(1)
	defer n.streamsOpen.Add(-1)
	in, out := pipe(s, c)
	n.bytesIn.Add(in)
	n.bytesOut.Add(out)
}

// Expose cria um listener TCP local cujas conexões viram streams para peerAddr.
// É o lado cliente: llama-server usa --rpc <local>.
func (n *Node) Expose(ctx context.Context, peerAddr, local string) (net.Listener, error) {
	ai, err := peer.AddrInfoFromString(peerAddr)
	if err != nil {
		return nil, err
	}
	n.Host.Peerstore().AddAddrs(ai.ID, ai.Addrs, time.Hour)
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err = n.Host.Connect(cctx, *ai)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("conectar a %s: %w", ai.ID, err)
	}
	ln, err := net.Listen("tcp", local)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				sctx, cancel := context.WithTimeout(ctx, 30*time.Second)
				s, err := n.Host.NewStream(sctx, ai.ID, RPCProtocol)
				cancel()
				if err != nil {
					slog.Error("abrir stream", "peer", ai.ID, "err", err)
					c.Close()
					return
				}
				pipe(c, s)
			}()
		}
	}()
	return ln, nil
}

type Stats struct {
	PeerID       string   `json:"peer_id"`
	Addrs        []string `json:"addrs"`
	Peers        int      `json:"connected_peers"`
	StreamsOpen  int64    `json:"streams_open"`
	StreamsTotal int64    `json:"streams_total"`
	BytesIn      int64    `json:"bytes_in"`
	BytesOut     int64    `json:"bytes_out"`
}

func (n *Node) Stats() Stats {
	var addrs []string
	for _, a := range n.Host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", a, n.Host.ID()))
	}
	return Stats{
		PeerID:       n.Host.ID().String(),
		Addrs:        addrs,
		Peers:        len(n.Host.Network().Peers()),
		StreamsOpen:  n.streamsOpen.Load(),
		StreamsTotal: n.streamsTotal.Load(),
		BytesIn:      n.bytesIn.Load(),
		BytesOut:     n.bytesOut.Load(),
	}
}

func IsRelayAddr(a multiaddr.Multiaddr) bool {
	_, err := a.ValueForProtocol(multiaddr.P_CIRCUIT)
	return err == nil
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

// pipe copia nos dois sentidos; retorna bytes a->b e b->a.
func pipe(a, b io.ReadWriteCloser) (int64, int64) {
	var wg sync.WaitGroup
	var ab, ba int64
	wg.Add(2)
	go func() { defer wg.Done(); ab, _ = io.Copy(b, a); closeWrite(b) }()
	go func() { defer wg.Done(); ba, _ = io.Copy(a, b); closeWrite(a) }()
	wg.Wait()
	a.Close()
	b.Close()
	return ab, ba
}

func closeWrite(c io.Closer) {
	switch v := c.(type) {
	case *net.TCPConn:
		v.CloseWrite()
	case network.Stream:
		v.CloseWrite()
	default:
		c.Close()
	}
}
