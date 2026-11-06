# ADR 0029: macOS não é bloqueado pelo design

- Status: aprovado (2026-11-06)

## Contexto

A premissa 15 cita web, desktop (Windows e Linux) e mobile (Android e iOS). O
brief 4.14 deixou em aberto (pergunta 16) o suporte a macOS: não está nos
requisitos, mas o design não deve impedi-lo.

## Decisão

1. **Nada no design bloqueia o macOS.**
2. **SDK e exemplos para macOS quando houver demanda.**

## Alternativas

- **Suporte completo desde já** (SDK, exemplo e testes em macOS): custo sem
  consumidor.
- **Declarar o macOS fora de suporte:** contraria a intenção do brief de não
  impedi-lo.

## Consequências

- Os padrões já previstos valem para o macOS sem mudança: navegador do sistema
  com retorno por loopback (RFC 8252), PKCE, armazenamento seguro da plataforma
  (Keychain) (brief 4.14).
- Nenhum `examples/` nem runner de macOS no CI por enquanto, exceto os
  previstos para iOS (brief 4.20).
- Pedir suporte a macOS é questão de roadmap, não de nova decisão de design.
