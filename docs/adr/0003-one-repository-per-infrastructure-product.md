# ADR 0003: um repositório por produto de infraestrutura

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.1 previa um monorepo para a plataforma, com AuthN, AuthZ e serviços
compartilhados, incluindo `notifications/`. A pergunta 6 da seção 9 pergunta
quais outros serviços de infraestrutura existem e se vivem no mesmo monorepo.
O iam é o primeiro de uma família de produtos de infraestrutura (e-mail e
outros virão), consumidos por projetos finais que ainda não existem
([ADR 0002](0002-test-client-projects.md)).

## Decisão

1. **Cada produto de infraestrutura no próprio repositório.** O iam contém os
   dois produtos irmãos de identidade: AuthN e AuthZ.
2. **`notifications/` sai do iam.** Fica a porta de envio `shared/messaging`,
   com adaptador SMTP direto e, depois, o adaptador do serviço de e-mail.
3. **Código comum é publicado como módulo Go** (por exemplo,
   `shared/statemachine`, [ADR 0022](0022-authz-states-and-transitions.md)).

## Alternativas

- **Monorepo único da "plataforma"** com todos os produtos de infraestrutura:
  acopla CI, releases e permissões de produtos sem relação entre si.
- **Monorepo só de AuthN e AuthZ, mantendo dentro dele os serviços que outros
  produtos também usam** (como `notifications/` no brief 4.1): os demais
  produtos passariam a depender do repositório de identidade para algo que não
  é identidade.

## Consequências

- Muda o brief 4.1: sem `notifications/`; entra `cmd/iam/` (design, seção 13).
- Produtos se integram por APIs públicas e módulos Go versionados, nunca por
  pacotes internos.
- Módulos publicados têm `go.mod` próprio e não arrastam dependências de
  servidor (design, seção 4.2).
- Até existir o serviço de e-mail, o envio é SMTP direto pela porta.
