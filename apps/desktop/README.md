# apps/desktop

Cliente desktop. Electron + React + TypeScript + Ant Design.

Só UI. Toda lógica de rede e inferência fica no daemon (`services/daemon`), acessado via gRPC em `localhost`.

Telas do MVP: onboarding (3 passos), painel (estado, pausar/retomar, stats), chat de inferência, configurações (limites, janelas horárias).

Bootstrap sugerido: `pnpm create electron-vite` (template react-ts).
