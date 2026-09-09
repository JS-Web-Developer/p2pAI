# packages/proto

Contratos protobuf.

- `daemon.proto` — daemon <-> UI (gRPC local): status, pausar/retomar, stats, configuração, inferência.
- `peer.proto` — nó <-> nó: handshake, capacidades, stream de ativações, recibos.
- `coordinator.proto` — nó/cliente <-> coordinator: registro, heartbeat, lista de nós, envio de recibos.

Geração: `buf generate` (Go e TypeScript).
