# ADR 0026: sem app autenticador próprio

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.14 deixou em aberto (pergunta 15) um app autenticador próprio para
MFA por push e para aprovações multipartes, ou um SDK mobile para os apps
clientes receberem essas notificações. Passkeys e chaves físicas (WebAuthn)
são o nível resistente a phishing; push com confirmação de número é
intermediário (brief 4.6). As aprovações M de N já usam assinatura WebAuthn
([ADR 0021](0021-authz-multi-party-approval.md)).

## Decisão

1. **Nenhum app autenticador próprio.**
2. Push e aprovações usam **passkeys/WebAuthn**.
3. **SDK mobile** para os apps clientes receberem notificações, em **camada
   posterior**.

## Alternativas

- **App autenticador próprio:** dois apps nativos a publicar e manter nas lojas,
  com superfície de segurança própria, para um fator abaixo do WebAuthn.

## Consequências

- Não há app nas lojas para manter, nem verificação MASVS de um app próprio.
- A confirmação forte de ações (step-up, M de N) é sempre WebAuthn com
  verificação de usuário.
- O suporte a passkeys de plataforma no Linux é limitado; chaves físicas USB
  são a alternativa (brief 4.14).
- Notificações a aprovadores em apps clientes dependem do SDK mobile da camada
  posterior.
