# ADR 0025: escopo do AuthZ no marco 1

- Status: aprovado (2026-11-06)
- Escopo: marco 1 ([ADR 0004](0004-milestones-0-and-1.md)).

## Contexto

O marco 1 exige que o cliente A consulte uma permissão RBAC via AuthZEN e que a
plataforma use esse RBAC nos papéis administrativos
([ADR 0004](0004-milestones-0-and-1.md)). O brief 4.29 quer a plataforma como
primeiro cliente do AuthZ, começando com RBAC simples já atrás do contrato que
a autorização completa vai implementar.

## Decisão

**No marco 1:**

| Item | ADR |
|---|---|
| Armazém de relações | [0013](0013-authz-hybrid-model.md), [0014](0014-authz-own-engine-with-oracle.md) |
| Políticas no núcleo Cedar, com versões validadas | [0015](0015-authz-cedar-core-language.md), [0019](0019-authz-policy-lifecycle.md) |
| Papéis por organização | [0013](0013-authz-hybrid-model.md) |
| AuthZEN: avaliação individual e em lote | [0024](0024-authz-standards.md) |
| Explicação básica | [0019](0019-authz-policy-lifecycle.md) |
| Auditoria de decisões | [0005](0005-state-tables-and-hash-chained-audit.md) |
| SDK Go com cache de decisões | [0017](0017-authz-local-evaluation-and-cache.md) |

A própria plataforma usa esse RBAC nos papéis administrativos: operador,
administrador de tenant e administrador de organização.

**No marco 2:** plano de filtro, avaliação local, outbox e tokens de
consistência ([ADR 0016](0016-authz-relation-sync-and-consistency.md),
[ADR 0017](0017-authz-local-evaluation-and-cache.md),
[ADR 0018](0018-authz-listing.md)).

## Alternativas

- **Marco 1 com plano de filtro, avaliação local, outbox e tokens de
  consistência:** o critério do marco 1 não precisa deles; foram para o
  marco 2.

## Consequências

- Os testes diferenciais, de propriedade, fuzzing e fail-closed do
  [ADR 0014](0014-authz-own-engine-with-oracle.md) valem desde o marco 1.
- Testes de abuso do marco 1 (design, seção 11): escalada de administrador de
  tenant para operador; decisão de AuthZ entre tenants; erro no meio da
  avaliação nega.
- Sem tokens de consistência, revogação chega ao SDK com atraso de até um TTL
  do cache: o TTL padrão precisa ser curto.
- O contrato AuthZEN (C7) é o mesmo que as camadas seguintes estendem.
- Acesso dinâmico, M de N, estados e agentes ficam fora
  ([ADR 0020](0020-authz-dynamic-access.md) a
  [ADR 0023](0023-authz-machines-and-agents.md)).
