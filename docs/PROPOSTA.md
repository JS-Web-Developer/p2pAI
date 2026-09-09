# P2PAI — Proposta de Projeto

Complementa o PRD em `readme.md`. Preenche: correções de arquitetura, estrutura do repositório, escopo do MVP, roadmap e seção de segurança (inacabada no PRD).

## 1. Correções ao PRD

| Ponto do PRD | Problema | Decisão |
|---|---|---|
| Petals + llama.cpp | Petals executa blocos como módulos PyTorch sobre hivemind. Trocar o backend é reescrever o núcleo. | Usar `ggml-rpc` (`rpc-server`) do llama.cpp, que já divide camadas entre hosts CPU/Metal/CUDA. Ver ADR-001. |
| Tailscale + libp2p | Tailscale é produto, não biblioteca. libp2p já faz NAT traversal (AutoNAT, Circuit Relay v2, DCUtR). | Só libp2p. |
| "P2P puro, sem servidor" | Mapa, placar e créditos exigem estado agregado. | Coordinator leve (bootstrap, relay, stats). DHT apenas para descoberta. |
| Token on-chain no MVP | Risco regulatório, distração. | Créditos off-chain. Token é decisão de Fase 3. |
| Mobile contribuindo | Bateria, térmica, upload residencial, políticas de loja. | Mobile só consome inferência. |
| Electron + subprocesso Python | Bundle pesado, distribuição frágil. | Electron só para UI. Daemon em Go, binário único, empacota `rpc-server`. Ver ADR-002. |

## 2. Estrutura do repositório

```
p2pai/
  apps/desktop/          Electron + React + TypeScript + Ant Design. Só UI; fala com o daemon via gRPC local.
  apps/web/              Next.js: site, mapa global, placar.
  services/daemon/       Go: host libp2p, scheduler de recursos, ponte para llama.cpp rpc-server.
  services/coordinator/  Go: bootstrap + relay + API de stats/créditos (Postgres).
  packages/proto/        Protobuf: daemon<->UI, nó<->nó, nó<->coordinator.
  spikes/fase0/          Experimentos descartáveis da Fase 0.
  docs/                  PRD, proposta, ADRs.
```

## 3. Arquitetura

### Papéis
- **Cliente**: pede inferência. Roda embeddings e amostragem localmente.
- **Nó**: contribui computação. Roda uma fatia contígua de camadas no `rpc-server`.
- **Coordinator**: bootstrap libp2p, relay, registro de nós, stats, créditos.

### Fluxo de inferência
1. Cliente consulta o coordinator: modelo desejado.
2. Coordinator responde lista de nós por tier (VRAM/RAM, largura de banda, reputação, uptime).
3. Cliente monta pipeline via libp2p (QUIC), abre stream para cada nó.
4. Cada nó executa sua fatia de camadas e passa ativações ao próximo.
5. Cliente amostra o token. Ao final, assina um recibo (tokens processados por nó) e envia ao coordinator.
6. Coordinator credita cada nó.

### Tiers de nó
| Tier | Hardware | Camadas (Llama 3.1 8B Q4) |
|---|---|---|
| 1 | GPU alta (4090/A100) | modelo inteiro ou 16+ |
| 2 | GPU média (3060/3070) | 8–16 |
| 3 | Mac série M | 4–8 |
| 4 | CPU AVX2 | 1–4 |

### Scheduler de recursos (daemon)
Pausa quando: CPU ou GPU > 80% por 10 s; processo de jogo detectado (Steam, Epic, Battle.net, Riot); janela manual de bloqueio.
Retoma quando: idle > 5 min, tela bloqueada, ou janela horária configurada.
Sempre disponível: pausar/retomar manual, limite de CPU/GPU/RAM.

## 4. Escopo do MVP

Histórias do PRD cobertas: 1, 2, 3, 4, 8, 10, 11, 14.

Entregas:
- Instalador Mac (Apple Silicon), Windows, Linux. Um clique, começa a contribuir.
- Pausa/retomada manual e automática.
- Stats locais: uptime, tokens processados, requisições atendidas.
- Inferência com **um** modelo (Llama 3.1 8B Q4_K_M) dividido em 2–4 nós.
- Chat simples no cliente com progresso e latência por nó.
- Onboarding em 3 telas.
- Rede fechada, entrada por convite.

Fora do MVP: mapa, placar, créditos, prioridade, votação, emblemas, API pública, web, mobile, token.

## 5. Roadmap

| Fase | Duração | Objetivo | Go/no-go |
|---|---|---|---|
| 0 — Spike | 3 semanas | 3 máquinas (Mac M, Linux CPU, NVIDIA) rodando `rpc-server` sobre stream libp2p em WAN real. | ≥ 1 token/s em 8B com 3 nós. Abaixo disso, repensar. |
| 1 — MVP | 8–10 semanas | Daemon + Electron + coordinator com 2 bootstrap nodes. Rede fechada. | 20 nós, 7 dias de uptime sem intervenção. |
| 2 — Comunidade | 6–8 semanas | Rede pública, mapa, placar, créditos, prioridade, onboarding completo. | 200 nós ativos. |
| 3 — Plataforma | contínuo | API pública, cliente web (consumo), reputação e verificação, decisão sobre token, mais modelos. | — |

## 6. Segurança e privacidade (completa a seção 7 do PRD)

- **Identidade**: cada nó tem par de chaves libp2p. Resultados e recibos assinados.
- **Verificação**: cliente reexecuta 5% das requisições em um segundo nó e compara ativações. Divergência derruba reputação de ambos até desempate por terceiro nó.
- **Reputação**: score por uptime, latência, taxa de divergência. Nós abaixo do limiar saem da lista do coordinator.
- **Privacidade**: nós veem ativações, não texto, mas ativações podem ser invertidas. O cliente exibe aviso claro: "não envie dados sensíveis". Modo privado (Fase 3): pipeline só com nós de confiança.
- **Abuso**: rate limit por peer no coordinator; recibos exigem assinatura do cliente; créditos só por requisições com verificação amostral aprovada.
- **Honeypots** (item original do PRD): coordinator injeta requisições com resultado conhecido para detectar nós que fabricam saída.
- **Cliente**: binários assinados (notarização Apple, Authenticode). Auto-update com assinatura verificada.

## 7. Métricas de sucesso

- Tempo entre download e primeira contribuição < 3 min.
- Tokens/s por modelo e por tier.
- Uptime médio de nó.
- Taxa de divergência na verificação < 1%.
- Retenção de nós em 30 dias.

## 8. Primeiros passos

1. Instalar toolchain: Go 1.22+, Node 20+, pnpm, llama.cpp (`brew install llama.cpp`).
2. Rodar spike da Fase 0 (`spikes/fase0/README.md`).
3. Registrar resultados em `docs/adr/ADR-003-resultado-spike.md`.
