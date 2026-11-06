# ADR 0018: plano de filtro e listagem paginada

- Status: aprovado (2026-11-06)
- Escopo: marco 2 ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

Listar "quais recursos este usuário pode acessar" é o ponto fraco dos motores
Zanzibar: o OpenFGA trunca em 3 s ou 1.000 itens, e Cerbos e SpiceDB tiveram
listagem mais permissiva que o check (`docs/market-study/summary.md`, seções 3
e 4). Os recursos dos clientes vivem nos bancos deles, e copiá-los para o AuthZ
seria escrita dupla ([ADR 0016](0016-authz-relation-sync-and-consistency.md)).

## Decisão

1. **Plano de filtro:** avaliação parcial gera um predicado estruturado, que o
   SDK traduz para SQL na consulta da própria aplicação.
2. **Listagem paginada por cursor no AuthZ** para o que depende só das relações
   dele, **sem truncar**.
3. **Índice materializado só se benchmark exigir.**
4. **Propriedade testada no CI:** listagem nunca mais permissiva que o check, e
   vice-versa.

## Alternativas

- **Só listagem no servidor:** exigiria copiar os recursos das aplicações para
  o AuthZ ou truncar resultados.
- **Índice materializado (estilo Leopard) desde já:** custo de escrita e de
  armazenamento sem medição que o justifique; fica condicionado a benchmark.

## Consequências

- Testes de propriedade obrigatórios: cada item listado passa no check, e cada
  recurso que passa no check aparece na listagem.
- Nunca trunca em silêncio; paginação por cursor (busca do AuthZEN,
  [ADR 0024](0024-authz-standards.md)).
- O tradutor de predicado para SQL no SDK é código crítico: testes contra
  Postgres real e nenhum valor concatenado na consulta.
- O avaliador precisa de avaliação parcial
  ([ADR 0014](0014-authz-own-engine-with-oracle.md)).
