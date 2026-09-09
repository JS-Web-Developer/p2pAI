# ADR-001: Inferência distribuída via llama.cpp RPC em vez de fork do Petals

**Status**: aceito
**Data**: 2026-09-09

## Contexto
O PRD propõe construir sobre o Petals e "substituir o backend de inferência" por llama.cpp. Petals executa blocos transformer como módulos PyTorch e coordena via hivemind (DHT própria). O backend não é plugável; substituí-lo equivale a reescrever o núcleo mantendo apenas a ideia.

llama.cpp possui `ggml-rpc`: um `rpc-server` que expõe um backend ggml pela rede. O `llama-cli`/`llama-server` pode distribuir camadas entre vários `rpc-server` (CPU, Metal, CUDA) com `--rpc host:port,...`.

## Decisão
Usar llama.cpp com `ggml-rpc` como motor de execução por nó. O transporte entre nós passa a ser stream libp2p em vez de TCP direto. Petals fica como referência de design (particionamento, fault tolerance), não como dependência.

## Consequências
- Suporte a CPU, Metal e CUDA sem código de backend próprio.
- Modelos em GGUF, quantizados, menor tráfego.
- Precisa de ponte: stream libp2p <-> socket local do `rpc-server`.
- `ggml-rpc` não tem autenticação nem tolerância a falha. Ambas ficam na camada do daemon.
- Risco: protocolo RPC do ggml muda sem garantia de compatibilidade. Fixar versão do llama.cpp por release.
