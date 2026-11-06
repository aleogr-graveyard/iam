# ADR 0011: metas de latência e de escala

- Status: aprovado (2026-11-06)

## Contexto

A premissa 22 pede alto desempenho, com orçamentos de latência por operação
verificados continuamente; o brief 4.21 propõe benchmarks no CI com detecção
de regressão. Faltavam números.

## Decisão

Metas de projeto, **verificadas por benchmark**. Não são promessa de
disponibilidade.

**Latência (p99 no servidor)**

| Operação | Meta |
|---|---|
| Validação de token no SDK | < 1 ms |
| Checagem AuthZEN (RBAC) | < 10 ms |
| Avaliação local | < 1 ms |
| Emissão de token (chave em memória) | < 50 ms |
| Login com passkey | < 100 ms |
| Login com senha | < 500 ms |

**Escala**

| Dimensão | Meta |
|---|---|
| Usuários por instalação | 10 milhões |
| Tenants | 100 mil |
| Organizações por tenant | 100 mil |
| Logins | 1.000/s |
| Checagens de autorização | 50.000/s |

## Alternativas

Nenhuma. As metas são objetivos de projeto verificados por benchmark, não
compromissos de disponibilidade; metas de disponibilidade, RPO e RTO ficam para
a saída do laboratório (design, seção 14).

## Consequências

- Benchmarks com detecção de regressão na rodada noturna (design, seção 11).
- O marco 1 mede as metas aplicáveis por benchmark
  ([ADR 0004](0004-milestones-0-and-1.md)).
- O login com senha inclui o Argon2id, lento por design: parâmetros calibrados,
  rate limiting e controle de concorrência (brief 4.21).
- A avaliação local depende do [ADR 0017](0017-authz-local-evaluation-and-cache.md).
- As metas de escala orientam modelo de dados, índices e paginação desde o
  início. O laboratório ([ADR 0001](0001-lab-hosting.md)) não é ambiente de
  medição de escala.
