# ADR 0022: estados e transições declarativos

- Status: aprovado (2026-11-06)
- Escopo: biblioteca interna desde o marco 1; funcionalidade do AuthZ em camada
  posterior ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

O brief 4.29 descreve três níveis: (1) biblioteca interna de máquinas de
estados; (2) funcionalidade do AuthZ em que o cliente declara estados e
transições dos seus recursos; (3) motor de workflow, fora do escopo. No cenário
da clínica, um prontuário em rascunho é editável e, depois de assinado, só
aceita adendos. O estudo confirma o nível 2 como diferencial: nenhum produto de
AuthZ trata estados e transições como conceito declarativo; todos usam estado
como atributo (`docs/market-study/authz/differentials-check.md`).

## Decisão

1. **Estados e transições declarativos**, com **definição única** usada pelo
   AuthZ para autorizar e pela aplicação (via SDK) para conduzir o fluxo.
2. A resposta do AuthZ traz as **ações e transições permitidas** (brief 4.34).
3. A aplicação é dona do estado e o informa a cada pedido; o AuthZ não
   sincroniza estados de negócio.
4. Políticas por transição, inclusive com M de N
   ([ADR 0021](0021-authz-multi-party-approval.md)).
5. `shared/statemachine` (nível 1) usa o mesmo modelo e é publicado como
   módulo Go próprio ([ADR 0003](0003-one-repository-per-infrastructure-product.md)).

## Alternativas

- **Estado só como atributo e transição como ação nomeada** (como todo o
  mercado): expressável, mas sem checar se a transição existe a partir do
  estado atual e com a definição duplicada na aplicação.
- **Motor de workflow** (nível 3): seria outro produto.

## Consequências

- Testes baseados em modelo: todas as transições válidas aceitas, todas as
  inválidas recusadas (brief 4.19).
- Diagramas de estados gerados a partir das definições (brief 4.4).
- A API da biblioteca é o contrato C5 (design, seção 8).
- O diferencial pode ser copiado; o valor está na definição única e na
  combinação "transição exige M de N assinado".
- As extensões de estado da linguagem se traduzem para o núcleo Cedar
  ([ADR 0015](0015-authz-cedar-core-language.md)).
