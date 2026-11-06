# ADR 0016: escrita de relações, outbox e tokens de consistência

- Status: aprovado (2026-11-06)
- Escopo: marco 2 ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

Com relações armazenadas no AuthZ ([ADR 0013](0013-authz-hybrid-model.md)),
dois problemas do brief 4.29 aparecem: a escrita dupla (a aplicação muda o
próprio banco e o AuthZ) e o "new enemy problem" do Zanzibar (decisão tomada
com dados anteriores a uma revogação). Nenhum projeto aberto resolve a escrita
dupla (`docs/market-study/summary.md`, seção 2). O laboratório não permite
`LISTEN/NOTIFY` ([ADR 0001](0001-lab-hosting.md)).

## Decisão

1. **API de escrita de relações.**
2. **Outbox transacional no SDK:** a aplicação grava a mudança de relação na
   própria transação; um **relay idempotente** a entrega ao AuthZ.
3. **Tokens de consistência** ("pelo menos tão atual quanto") para revogações
   sensíveis.
4. **Consistência eventual configurável** no resto.
5. **Modo embutido na mesma transação (Go)** como camada posterior possível.
6. Consulta sob demanda e CDC só com demanda real.

## Alternativas

- **Modo embutido na mesma transação:** elimina a escrita dupla, mas só para
  aplicações Go no mesmo banco; fica como camada posterior.
- **Consulta sob demanda** (o AuthZ busca dados na aplicação a cada decisão):
  acopla latência e disponibilidade das duas.
- **CDC ou eventos:** depende do banco e da infraestrutura de cada aplicação.

## Consequências

- O relay é idempotente: entregar a mesma mudança mais de uma vez não altera o
  resultado (teste obrigatório).
- Outbox lido por polling ([ADR 0001](0001-lab-hosting.md)).
- O token de consistência invalida o cache do SDK
  ([ADR 0017](0017-authz-local-evaluation-and-cache.md)).
- O AuthZEN 1.0 não define token de consistência; ele precisa de
  representação própria, documentada.
- Teste obrigatório: decisão pedida com o token de uma revogação nunca usa
  dados anteriores a ela.
