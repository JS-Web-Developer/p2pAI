// Package scheduler: decide se o nó contribui agora.
//
// Pausa: CPU/GPU > 80% por 10 s, processo de jogo, janela manual.
// Retoma: idle > 5 min, tela bloqueada, janela horária.
package scheduler
