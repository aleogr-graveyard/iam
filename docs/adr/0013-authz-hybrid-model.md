# ADR 0013: AuthZ com modelo híbrido (relações + políticas)

- Status: aprovado (2026-11-06)

## Contexto

O AuthZ deve cobrir RBAC, ABAC, ReBAC e PBAC, combináveis (premissa 1,
brief 4.29 e 5.2). O cenário da clínica mistura os três tipos de regra: a
Dra. Ana vê só os prontuários dos pacientes que atende (relação), a recepção vê
a agenda (papel) e um prontuário assinado só aceita adendos (atributo e
estado). No mercado, os motores fazem bem ReBAC ou ABAC, raramente os dois
(`docs/market-study/summary.md`, seção 1).

## Decisão

1. **Relações armazenadas no AuthZ**, por exemplo
   `organização#membro@usuário` e `paciente#atendido_por@usuário`.
2. **Políticas `permit`/`forbid` com condições** que consultam relações,
   atributos e contexto.
3. **RBAC é caso particular**: papel por organização é uma relação.

## Alternativas

- **Zanzibar puro** (só relações, estilo OpenFGA e SpiceDB): condições sobre
  atributos e contexto ficam como acessório do modelo.
- **Políticas sem estado puras** (estilo Cedar e Cerbos): a aplicação envia
  todas as entidades e relações a cada pedido, e a listagem fica com ela.

## Consequências

- O AuthZ precisa de armazém de relações e de avaliador de políticas, com
  decisões coerentes entre si ([ADR 0014](0014-authz-own-engine-with-oracle.md)).
- A linguagem precisa ler relações ([ADR 0015](0015-authz-cedar-core-language.md)).
- A consistência das relações passa a ser problema do AuthZ
  ([ADR 0016](0016-authz-relation-sync-and-consistency.md)).
- A listagem precisa combinar relações e condições
  ([ADR 0018](0018-authz-listing.md)).
- Os papéis por organização do marco 1 já usam este modelo
  ([ADR 0025](0025-authz-milestone-1-scope.md)).
