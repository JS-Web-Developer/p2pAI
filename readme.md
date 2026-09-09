# P2PAI - Documento de Requisitos do Produto (PRD)

## Definição do Problema

O uso atual de Grandes Modelos de Linguagem (LLMs) enfrenta dois desafios fundamentais:

1. **Alta Barreira de Entrada**: Para usuários comuns, acessar grandes modelos (como Llama 3, DeepSeek, etc.) exige hardware de GPU caro ou taxas elevadas de uso de API, limitando a adoção em massa e a democratização da tecnologia de IA.

2. **Desperdício de Recursos Computacionais**: Centenas de milhões de computadores pessoais, Macs e até dispositivos móveis em todo o mundo permanecem ociosos, com seu poder computacional subutilizado. Enquanto isso, projetos existentes de computação distribuída para IA (como o Petals) visam principalmente usuários técnicos, criando uma barreira muito alta para o público em geral.

Da perspectiva do usuário, o problema pode ser descrito da seguinte forma:
- **Como entusiasta de IA**: Quero usar grandes modelos, mas não quero pagar taxas exorbitantes nem comprar hardware de GPU caro.
- **Como colaborador para o bem público**: Quero doar meu poder computacional ocioso para ajudar outras pessoas, mas as ferramentas existentes são muito complexas e não sei como começar.
- **Como membro da comunidade de código aberto**: Quero uma plataforma simples e intuitiva que me permita participar facilmente do movimento de democratização da IA.

## Solução

O **P2PAI** é uma plataforma descentralizada de compartilhamento de recursos computacionais para IA, baseada no protocolo Petals, que permite a usuários de todo o mundo contribuir facilmente com poder computacional ocioso e acessar serviços de grandes modelos.

### Principais Soluções

1. **Cliente sem Barreiras**: Desenvolvimento de um cliente desktop semelhante ao Folding@home; os usuários podem começar a contribuir com poder computacional por meio de uma instalação com um único clique, sem necessidade de configurações complexas.

2. **Suporte Multiplataforma**: Por meio da integração com o `llama.cpp`, a plataforma oferece suporte a GPUs NVIDIA, CPUs, Macs (Apple Silicon) e até dispositivos móveis, permitindo a participação de uma ampla gama de hardwares.

3. **Agendamento Inteligente de Recursos**: Detecção automática da carga do sistema para pausar operações quando o usuário estiver jogando ou trabalhando e retomá-las quando o sistema estiver ocioso, possibilitando uma "contribuição sem atrito".

4. **Mecanismo de Incentivo via Tokens**: Introdução de um token utilitário no qual os usuários ganham recompensas por contribuir com poder computacional; esses tokens podem ser trocados por benefícios, como acesso prioritário a inferências, estabelecendo um modelo de incentivo sustentável. 5. **Comunidade Visualizada**: Apresenta um mapa global de computação, placares de líderes de contribuição, emblemas da comunidade e muito mais, permitindo que os usuários visualizem o valor de suas contribuições e promovendo um senso de pertencimento à comunidade.

### Vantagens Diferenciais

Comparado às soluções existentes (por exemplo, Petals, Hugging Face):

| Dimensão | Petals | Hugging Face | P2PAI |
|-----------|--------|--------------|-------|
| Barreira para o Usuário | Alta (Linha de Comando) | Média (Interface Web) | Baixa (Instalação com um clique) |
| Suporte de Plataforma | Principalmente GPU | GPU | GPU + CPU + Mac + Mobile |
| Mecanismo de Incentivo | Nenhum | Nenhum | Tokens + Sistema de Honra |
| Agendamento de Recursos | Manual | Manual | Inteligente/Automatizado |
| Visualização da Comunidade | Nenhuma | Básica | Mapa Global + Placares de Líderes |

## Histórias de Usuário

### Histórias de Usuário Principais (Essenciais para o MVP)

#### Contribuidores de Computação

1. Como usuário de Mac, quero baixar um cliente de desktop que comece a contribuir com poder computacional automaticamente após uma instalação com um clique, para que eu possa participar desta iniciativa de bem público sem aprender operações complexas de linha de comando.

2. Como jogador de PC, quero que o cliente detecte de forma inteligente quando estou jogando e pause automaticamente as contribuições de computação, garantindo que minha experiência de jogo não seja afetada.

3. Como usuário de computador de escritório, quero pausar manualmente as contribuições durante o horário de trabalho e fazer com que elas sejam retomadas automaticamente após o expediente, garantindo que minha produtividade não seja impactada.

4. Como contribuidor de computação, quero ver a quantidade de poder computacional que contribuí e o número de tarefas de inferência que apoiei dentro do cliente, para que eu possa sentir o valor da minha contribuição.

5. Como contribuidor de computação, quero ver o número de nós ativos em todo o mundo em um mapa, para que eu possa perceber a escala e a vitalidade da comunidade.

6. Como um dos principais contribuidores, quero ver minha classificação no placar de líderes, para que eu possa obter uma sensação de realização e honra.

7. Como detentor de tokens, quero visualizar meu saldo de tokens e histórico de ganhos, para que eu possa acompanhar meus retornos. #### Usuários de Modelos

8. Como um estudante de IA, quero usar a rede compartilhada para inferência de modelos via cliente, para que eu possa aprender sobre modelos grandes sem comprar GPUs caras. 9. Como detentor de tokens, quero trocar tokens por direitos de prioridade na inferência, para que eu possa receber respostas rápidas mesmo durante períodos de pico.

10. Como usuário de modelos, quero escolher entre diferentes modelos (Llama, DeepSeek, etc.) para inferência, para que eu possa selecionar o modelo mais adequado às minhas necessidades.

11. Como usuário de modelos, quero visualizar o progresso e a latência da inferência, para que eu possa entender o status da rede em tempo real.

#### Membros da Comunidade

12. Como membro da comunidade, quero votar em qual modelo será executado na próxima semana, para que eu possa participar da tomada de decisões da comunidade.

13. Como membro da comunidade, quero ganhar medalhas de contribuição e emblemas de conquista, para que eu possa exibir minhas contribuições dentro da comunidade.

14. Como novo usuário, quero ver um tutorial de integração (*onboarding*), para que eu possa aprender rapidamente a usar a plataforma.

### Histórias de Usuário Secundárias (A serem implementadas em versões futuras)

#### Recursos Avançados

15. Como usuário avançado, quero personalizar as camadas do modelo e os limites de recursos que contribuo, para que eu possa ter um controle mais refinado sobre minha contribuição.

16. Como desenvolvedor, quero acessar a rede P2PAI via API, para que eu possa integrar serviços de inferência às minhas próprias aplicações.

17. Como usuário corporativo, quero criar uma rede de computação privada, para que minha equipe possa compartilhar poder computacional internamente.

#### Suporte para Dispositivos Móveis

18. Como usuário de dispositivo móvel, quero instalar um cliente no meu celular para contribuir com poder computacional, para que eu possa participar de uma iniciativa de bem público.

19. Como usuário de tablet, quero usar serviços de inferência no meu iPad, para que eu possa utilizar modelos grandes em um dispositivo móvel.

#### Suporte para Web

20. Como usuário casual, quero contribuir com poder computacional ou usar serviços de inferência diretamente via página web, para que eu não precise instalar nenhum software. 21. Como usuário da web, quero contribuir com poder computacional via WebGPU para poder utilizar os recursos de GPU do navegador.

#### Governança da Comunidade

22. Como participante da governança da comunidade, quero enviar propostas para suporte a novos modelos para impulsionar o desenvolvimento da comunidade.

23. Como detentor de tokens, quero participar das votações da comunidade para influenciar a direção do projeto. 24. Como operador de nó, quero visualizar estatísticas detalhadas do nó (tempo de atividade/uptime, volume de contribuição, pontuação de reputação, etc.) para otimizar minha estratégia de contribuição.

#### Segurança e Privacidade

25. Como usuário preocupado com a privacidade, quero saber como a plataforma protege meus dados para poder utilizá-la com confiança.

26. Como usuário preocupado com a segurança, quero ver os relatórios de auditoria de segurança da plataforma para poder confiar em sua segurança.

27. Como operador de nó, quero definir o nível de privacidade das minhas contribuições (público/anônimo) para controlar a extensão da divulgação das minhas informações.

#### Suporte a Múltiplos Idiomas

28. Como usuário que não fala inglês, quero que o cliente ofereça suporte ao meu idioma nativo para facilitar o uso.

29. Como usuário internacional, quero ver a distribuição global dos nós para entender o nível de internacionalização da comunidade.

#### Educação e Treinamento

30. Como iniciante, quero assistir a tutoriais em vídeo para começar mais rapidamente.

31. Como educador, quero usar o P2PAI para demonstrações de ensino e apresentar aos alunos os conceitos de IA distribuída.

32. Como pesquisador, quero acessar dados e estatísticas da rede para realizar pesquisas e análises.

## Decisões de Implementação

### Decisões de Arquitetura Técnica

#### 1. Desenvolvimento Secundário Baseado no Protocolo Petals

**Decisão**: Optar por construir sobre o protocolo Petals em vez de criar um novo protocolo do zero.

**Justificativa**:
- O Petals já implementou um mecanismo maduro de inferência distribuída P2P.
- A comunidade é ativa, contando com documentação e suporte existentes.
- Permite uma validação rápida de prova de conceito e um ciclo de desenvolvimento mais curto. **Detalhes de Implementação**:
- Manter a camada central de comunicação P2P e o mecanismo de particionamento de modelo do Petals.
- Substituir o *backend* de inferência e integrar o `llama.cpp` para oferecer suporte a ambientes de CPU e Mac.
- Desenvolver uma nova camada de cliente que ofereça interface gráfica de usuário e agendamento inteligente.

#### 2. Adaptação de *Backend* de Inferência Diversificado

**Decisão**: Viabilizar suporte multiplataforma integrando o `llama.cpp`. **Abordagem Técnica**:
- **Nós de GPU**: Continuar utilizando a configuração nativa do Petals (PyTorch + CUDA).
- **Nós de CPU**: Integrar o `llama.cpp` com suporte para conjuntos de instruções AVX2/AVX-512.
- **Nós Mac**: Utilizar o *backend* Metal do `llama.cpp` para otimizar o desempenho no Apple Silicon.
- **Dispositivos Móveis**: Utilizar a versão do `llama.cpp` otimizada para dispositivos móveis.

**Design da Interface**:
```python
class InferenceBackend(ABC):
@abstractmethod
def load_model_layer(self, layer_id: int, config: ModelConfig) -> None:
"""Carrega uma camada específica do modelo"""
pass

@abstractmethod
def forward(self, input_tensor: Tensor) -> Tensor:
"""Executa a inferência *forward*"""
pass

@abstractmethod
def get_device_info(self) -> DeviceInfo:
"""Retorna informações do dispositivo (VRAM/RAM, capacidade computacional, etc.)"""
pass
```

#### 3. Arquitetura de Rede P2P Pura

**Decisão**: Adotar uma rede P2P pura, sem servidores centralizados.

**Abordagem Técnica**:
- Utilizar `libp2p` para descoberta de nós e comunicação.
- Integrar o Tailscale (protocolo WireGuard) para travessia de NAT.
- Utilizar DHT (*Distributed Hash Table*) para armazenar informações dos nós e metadados do modelo.

**Topologia de Rede**:
```
Tipos de Nó:
- Supernós: Alta largura de banda e estabilidade; responsáveis ​​pelo roteamento e coordenação.
- Nós Padrão: Contribuem com poder computacional; participam da inferência.
- Nós Cliente: Apenas consomem serviços de inferência.

Protocolos de Comunicação:
- Descoberta de Nós: libp2p Kademlia DHT
- Transferência de Dados: libp2p QUIC (baseado em UDP; rápido e confiável)
- Travessia de NAT: Relay Tailscale DERP + conexão direta
```

#### 4. Agendamento de Pipeline Heterogêneo

**Decisão**: Alocar dinamicamente camadas do modelo com base na capacidade computacional do nó. **Estratégia de Agendamento**:
```
Hierarquia de Computação:
- Nível 1 (GPU de alto desempenho): RTX 4090/A100, etc.; processa 20–30 camadas
- Nível 2 (GPU intermediária): RTX 3060/3070, etc.; processa 10–15 camadas
- Nível 3 (Mac série M): M1/M2/M3; processa 2–5 camadas
- Nível 4 (CPU): CPU padrão; processa apenas camadas de amostragem ou camadas pequenas

Algoritmo de Agendamento:
1. Nós informam dados do dispositivo (modelo da GPU, VRAM, largura de banda, etc.) ao iniciar
2. O agendador atribui camadas do modelo com base nas informações do dispositivo
3. Múltiplos nós competem pela mesma camada; o nó que retorna o resultado mais rapidamente vence
4. Ajuste dinâmico: Tarefas são migradas automaticamente para outros nós se um nó específico responder lentamente
```

#### 5. Stack Tecnológica do Cliente Desktop

**Decisão**: Utilizar Electron + React + Ant Design.

**Justificativa**:
- Electron: Suporte multiplataforma (Windows/Mac/Linux) e desenvolvimento rápido
- React: Ecossistema maduro e amplas bibliotecas de componentes
- Ant Design: Componentes de interface (UI) de nível empresarial; esteticamente agradáveis ​​e profissionais

**Arquitetura do Cliente**:
```
Camada de Frontend (Processo de Renderização do Electron):
- React + TypeScript
- Biblioteca de componentes Ant Design
- Redux Toolkit para gerenciamento de estado
- React Query para busca de dados

Camada de Backend (Processo Principal do Electron):
- Node.js + subprocesso Python
- Comunicação com nós Petals
- Monitoramento de recursos do sistema
- Armazenamento de dados local

Camada de Comunicação:
- IPC (Comunicação entre Processos): Frontend para Backend
- WebSocket: Comunicação com a rede P2P
```

#### 6. Agendamento Inteligente de Recursos

**Decisão**: Implementar um modo híbrido (manual + automação inteligente). **Estratégia de Agendamento**:
```
Regras de Agendamento Automático:
1. Detecção de processos de jogos (Steam, Epic Games, etc.) → Pausa automática
2. Detecção de alto uso de CPU/GPU (>80%) → Pausa automática
3. Detecção de bloqueio de tela/protetor de tela → Início automático
4. Detecção de horário noturno (configurável) → Início automático

Controle Manual:
- Usuários podem pausar/retomar a contribuição manualmente
- Usuários podem definir intervalos de horário para contribuição
- Usuários podem definir limites de uso de recursos (utilização de CPU/GPU)
```

#### 7. Mecanismos de Segurança

**Decisão**: Mecanismo de proteção de segurança em múltiplas camadas.

**Medidas de Segurança**:
```
1. Honeypots
```
