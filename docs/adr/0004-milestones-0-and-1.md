# ADR 0004: marco 0 "fábrica" e marco 1 com AuthN e AuthZ mínimos

- Status: aprovado (2026-11-06)

## Contexto

Todo o código é escrito por agentes; o dono define, opera e aprova
(premissa 14). A verificação automatizada é o principal mecanismo de revisão.
O dono revisa **mais de 10 h por semana**: o gargalo não é a revisão humana, é
a qualidade da verificação automatizada. O exemplo de roadmap do brief 4.26
começava por "fundação compartilhada e AuthN mínimo".

## Decisão

1. **Marco 0 "fábrica"**, antes de qualquer código de produto:

| Entrega | Para quê |
|---|---|
| Estrutura de módulos e `go.work` | Base para fronteiras e publicação |
| CI geral, fronteiras (`depguard`), varreduras | Verificação automatizada confiável |
| Ferramentas de contrato (`buf`) | Contratos antes do código |
| Ambiente local | Um comando sobe Postgres e o servidor |
| Harness e2e com o cliente A | Teste ponta a ponta desde o início |
| Checks obrigatórios no ruleset | Nada entra sem passar |

2. **Marco 1: AuthN e AuthZ mínimos, ponta a ponta, no laboratório.** O
   cliente A faz login OIDC com senha e com passkey numa organização de um
   tenant, recebe token e consulta uma permissão RBAC via AuthZEN. A própria
   plataforma usa esse RBAC nos papéis administrativos.
3. Critérios de conclusão detalhados: design, seção 12.

## Alternativas

- **Marco 1 só com AuthN** (como no exemplo do brief 4.26): adia o contrato
  AuthZEN e deixa os papéis administrativos fora do AuthZ (brief 4.29).
- **Marco 1 só com AuthZ e IdP externo:** prova a premissa 8, mas adia o AuthN
  de que o cliente A precisa.
- **Planejar para uma capacidade de revisão menor:** marcos menores e menos
  frentes paralelas; descartado pela capacidade declarada.

## Consequências

- Nenhum código de produto antes de o marco 0 estar concluído.
- O marco 1 cruza os dois produtos: os contratos C7 (AuthZEN), C8 (OIDC) e C9
  (API de login guiada) precisam existir cedo (design, seção 8).
- Os testes de abuso da seção 11 do design fazem parte do critério do marco 1.
- Escopo do AuthZ no marco 1: [ADR 0025](0025-authz-milestone-1-scope.md);
  do AuthN: design, seção 6.
- Paralelização limitada pela capacidade de acompanhar e aprovar (brief 4.25).
