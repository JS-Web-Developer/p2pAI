// p2pai-daemon: nó da rede P2PAI.
//
// Responsabilidades:
//   - host libp2p (descoberta, relay, hole punching)
//   - scheduler de recursos (pausa/retomada)
//   - ponte stream libp2p <-> llama.cpp rpc-server local
//   - servidor gRPC local para a UI (packages/proto/daemon.proto)
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "p2pai-daemon: esqueleto. Ver docs/PROPOSTA.md e spikes/fase0.")
	os.Exit(1)
}
