# ADR-002: Daemon do nó em Go, Electron apenas para UI

**Status**: aceito
**Data**: 2026-09-09

## Contexto
O PRD propõe Electron com subprocesso Python. Python exige empacotar interpretador e dependências (PyTorch, hivemind) por plataforma; bundle de centenas de MB e updates frágeis. O daemon precisa de libp2p, monitor de recursos, scheduler e ponte para o `rpc-server`.

## Decisão
Daemon em Go, binário estático por plataforma, empacotado junto com o `rpc-server` do llama.cpp. Electron + React + Ant Design só para UI, comunicando com o daemon via gRPC em `localhost`.

## Justificativa
- `go-libp2p` é a implementação de referência: AutoNAT, Circuit Relay v2, DCUtR, QUIC prontos.
- Cross-compile trivial (darwin/arm64, windows/amd64, linux/amd64).
- Daemon roda sem UI (headless, servidores, CI).
- UI pode ser trocada (Tauri, web) sem tocar o núcleo.

## Consequências
- Dois toolchains (Go e Node) no repositório.
- Contrato daemon<->UI precisa ser definido em protobuf desde o início (`packages/proto`).
