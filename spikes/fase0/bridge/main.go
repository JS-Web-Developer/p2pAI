// bridge: ponte entre streams libp2p e o rpc-server do llama.cpp.
//
// Modos:
//
//	bridge relay  [-listen ma] [-key file]
//	    Host libp2p com serviço de relay (Circuit Relay v2). Papel do coordinator.
//
//	bridge node   -rpc 127.0.0.1:50052 [-listen ma] [-relay ma] [-key file]
//	    Aceita streams /p2pai/rpc/0.1.0 e encaminha ao rpc-server local.
//
//	bridge client -peer ma -local 127.0.0.1:60001 [-peer ma -local ...] [-relay ma]
//	    Para cada par (-peer, -local): listener TCP local; cada conexão abre um
//	    stream para o peer. Depois: llama-completion --rpc 127.0.0.1:60001,...
package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
)

const rpcProtocol protocol.ID = "/p2pai/rpc/0.1.0"

type multiFlag []string

func (m *multiFlag) String() string     { return fmt.Sprint([]string(*m)) }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	mode := os.Args[1]
	fs := flag.NewFlagSet(mode, flag.ExitOnError)

	listen := fs.String("listen", "", "multiaddr de escuta (padrão: QUIC e TCP em portas aleatórias)")
	keyFile := fs.String("key", "", "arquivo da chave privada (criado se não existir)")
	var relays, peers, locals multiFlag
	fs.Var(&relays, "relay", "multiaddr de relay estático (repetível)")
	rpcAddr := fs.String("rpc", "127.0.0.1:50052", "endereço do rpc-server local (modo node)")
	fs.Var(&peers, "peer", "multiaddr completo do nó remoto, com /p2p/<id> (repetível, modo client)")
	fs.Var(&locals, "local", "endereço TCP local para expor o peer correspondente (repetível, modo client)")
	fs.Parse(os.Args[2:])

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	h, err := newHost(mode, *listen, *keyFile, relays)
	if err != nil {
		fatal(err)
	}
	defer h.Close()
	printAddrs(h)

	switch mode {
	case "relay":
		if _, err := relay.New(h, relay.WithInfiniteLimits()); err != nil {
			fatal(err)
		}
		fmt.Println("relay ativo")
	case "node":
		h.SetStreamHandler(rpcProtocol, func(s network.Stream) {
			handleNodeStream(s, *rpcAddr)
		})
		fmt.Printf("encaminhando %s -> %s\n", rpcProtocol, *rpcAddr)
	case "client":
		if len(peers) == 0 || len(peers) != len(locals) {
			fatal(fmt.Errorf("-peer e -local devem aparecer em pares"))
		}
		for i := range peers {
			if err := serveClient(ctx, h, peers[i], locals[i]); err != nil {
				fatal(err)
			}
		}
	default:
		usage()
	}

	<-ctx.Done()
	fmt.Println("encerrando")
}

func newHost(mode, listen, keyFile string, relays []string) (host.Host, error) {
	priv, err := loadOrCreateKey(keyFile)
	if err != nil {
		return nil, err
	}
	addrs := []string{listen}
	if listen == "" {
		addrs = []string{
			"/ip4/0.0.0.0/udp/0/quic-v1",
			"/ip4/0.0.0.0/tcp/0",
		}
	}
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(addrs...),
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
	}
	if mode == "relay" {
		opts = append(opts, libp2p.EnableRelayService(), libp2p.ForceReachabilityPublic())
	} else {
		opts = append(opts, libp2p.EnableRelay())
		if len(relays) > 0 {
			var infos []peer.AddrInfo
			for _, r := range relays {
				ai, err := peer.AddrInfoFromString(r)
				if err != nil {
					return nil, fmt.Errorf("relay %q: %w", r, err)
				}
				infos = append(infos, *ai)
			}
			opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(infos))
		}
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	for _, r := range relays {
		ai, _ := peer.AddrInfoFromString(r)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := h.Connect(ctx, *ai); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: conectar ao relay %s: %v\n", ai.ID, err)
		} else {
			fmt.Printf("conectado ao relay %s\n", ai.ID)
		}
		cancel()
	}
	return h, nil
}

func loadOrCreateKey(path string) (crypto.PrivKey, error) {
	if path == "" {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		return priv, err
	}
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

func printAddrs(h host.Host) {
	fmt.Printf("peer id: %s\n", h.ID())
	for _, a := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", a, h.ID())
	}
	// Endereços via relay aparecem depois do AutoRelay negociar; imprimir periodicamente.
	go func() {
		seen := map[string]bool{}
		for range time.Tick(3 * time.Second) {
			for _, a := range h.Addrs() {
				s := a.String()
				if !seen[s] && isRelayAddr(a) {
					seen[s] = true
					fmt.Printf("  (relay) %s/p2p/%s\n", s, h.ID())
				}
			}
		}
	}()
}

func isRelayAddr(a multiaddr.Multiaddr) bool {
	_, err := a.ValueForProtocol(multiaddr.P_CIRCUIT)
	return err == nil
}

// node: stream libp2p -> TCP local do rpc-server.
func handleNodeStream(s network.Stream, rpcAddr string) {
	start := time.Now()
	c, err := net.DialTimeout("tcp", rpcAddr, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial %s: %v\n", rpcAddr, err)
		s.Reset()
		return
	}
	fmt.Printf("stream de %s (%s)\n", s.Conn().RemotePeer(), connKind(s.Conn()))
	n1, n2 := pipe(s, c)
	fmt.Printf("stream fechado: %d B recebidos, %d B enviados, %s\n", n1, n2, time.Since(start).Round(time.Millisecond))
}

// client: TCP local -> stream libp2p para o peer.
func serveClient(ctx context.Context, h host.Host, peerAddr, local string) error {
	ai, err := peer.AddrInfoFromString(peerAddr)
	if err != nil {
		return fmt.Errorf("peer %q: %w", peerAddr, err)
	}
	h.Peerstore().AddAddrs(ai.ID, ai.Addrs, time.Hour)

	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t := time.Now()
	err = h.Connect(cctx, *ai)
	cancel()
	if err != nil {
		return fmt.Errorf("conectar a %s: %w", ai.ID, err)
	}
	kind := "?"
	for _, c := range h.Network().ConnsToPeer(ai.ID) {
		kind = connKind(c)
	}
	fmt.Printf("conectado a %s em %s via %s\n", ai.ID, time.Since(t).Round(time.Millisecond), kind)

	ln, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}
	fmt.Printf("  %s -> %s\n", local, ai.ID)
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
				s, err := h.NewStream(sctx, ai.ID, rpcProtocol)
				cancel()
				if err != nil {
					fmt.Fprintf(os.Stderr, "stream para %s: %v\n", ai.ID, err)
					c.Close()
					return
				}
				pipe(c, s)
			}()
		}
	}()
	return nil
}

func connKind(c network.Conn) string {
	if isRelayAddr(c.RemoteMultiaddr()) {
		return "relay"
	}
	return "direta " + c.RemoteMultiaddr().String()
}

// pipe copia nos dois sentidos e fecha ambos ao fim. Retorna bytes a->b, b->a.
func pipe(a io.ReadWriteCloser, b io.ReadWriteCloser) (int64, int64) {
	var wg sync.WaitGroup
	var ab, ba int64
	wg.Add(2)
	go func() {
		defer wg.Done()
		ab, _ = io.Copy(b, a)
		closeWrite(b)
	}()
	go func() {
		defer wg.Done()
		ba, _ = io.Copy(a, b)
		closeWrite(a)
	}()
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

func usage() {
	fmt.Fprintln(os.Stderr, "uso: bridge relay|node|client [flags]  (-h para flags)")
	os.Exit(2)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "erro:", err)
	os.Exit(1)
}
