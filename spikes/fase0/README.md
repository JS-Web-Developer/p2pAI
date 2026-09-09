# Spike Fase 0 — llama.cpp RPC sobre libp2p em WAN

## Objetivo
Validar viabilidade: Llama 3.1 8B Q4_K_M dividido em 3 máquinas heterogêneas, em redes residenciais distintas, transporte via libp2p.

Go/no-go: ≥ 1 token/s.

## Passo 1 — RPC puro, LAN (sem libp2p)
Em cada nó:
```sh
ggml-rpc-server -H 0.0.0.0 -p 50052
```
No cliente:
```sh
llama-completion -m llama-3.1-8b-q4_k_m.gguf \
  --rpc nodeA:50052,nodeB:50052,nodeC:50052 \
  -ngl 99 -p "Explique computação distribuída em 3 frases" -n 128
```
Registrar: tokens/s, distribuição de camadas por nó (log do llama.cpp).

## Passo 2 — RPC sobre libp2p, WAN
Programa Go mínimo (`bridge/`): host libp2p com relay e hole punching, protocolo `/p2pai/rpc/0.1.0`. No nó, cada stream recebido é conectado a `localhost:50052`. No cliente, listener TCP local por nó remoto que abre stream para o peer. `llama-completion --rpc 127.0.0.1:60001,127.0.0.1:60002,...`.

Registrar: tokens/s, latência de handshake, se houve conexão direta ou relay.

## Passo 3 — Falha de nó
Matar um `rpc-server` no meio da geração. Observar comportamento do llama.cpp (esperado: aborta). Anotar o que o daemon precisa fazer para reconectar/refazer pipeline.

## Resultado
Escrever `docs/adr/ADR-003-resultado-spike.md` com números e decisão.

## Compilação (pendente)

llama.cpp do Homebrew não inclui RPC. Compilar do fonte (já clonado em `llama.cpp/`, ignorado pelo git):
```sh
cd spikes/fase0/llama.cpp
cmake -B build -DGGML_RPC=ON -DGGML_METAL=ON -DLLAMA_CURL=OFF -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release -j 8 --target ggml-rpc-server llama-completion llama-server llama-bench
```
Binários em `build/bin/`. Modelo de teste já baixado em `models/qwen2.5-1.5b-instruct-q4_k_m.gguf` (M1 8 GB não comporta 8B local; 8B fica para o teste em WAN).

Ponte libp2p: `bridge/` (compila com `go build`).
