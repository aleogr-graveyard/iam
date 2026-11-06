# ADR 0014: AuthZ com motor próprio e `cedar-go` como oráculo de teste

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.29 deixou em aberto (pergunta 8) motor próprio, motor existente
embutido ou motor existente como serviço. Pelo estudo de mercado
(`docs/market-study/summary.md`, seção 5; `authz/cedar.md`; `authz/openfga.md`):

- nenhum motor cobre ReBAC + ABAC + estados + listagem + consistência;
- fornecedores de AuthZ são instáveis (Aserto encerrado, Oso arquivado,
  Permify relicenciado);
- a coerência da autorização é a classe de falha mais comum (29 advisories no
  OpenFGA; listagem mais permissiva que o check no Cerbos e no SpiceDB).

A premissa 24 pede o mínimo de dependências; o avaliador é código crítico
(brief 4.25).

## Decisão

1. **Motor próprio:** armazém de relações no Postgres atrás de uma porta,
   avaliador e listagem próprios (`authz/core`). Nenhuma dependência de motor
   no runtime.
2. **`cedar-go` só em teste**, como oráculo dos testes diferenciais.
3. **Obrigatórios:** testes diferenciais, fuzzing, testes de propriedade
   (check, lote e listagem coerentes) e auditoria externa antes de produção.

## Alternativas

- **`cedar-go` embutido no runtime:** 3 dependências e muito rápido, mas sem
  avaliação parcial (base da listagem), com lote e validação fora do semver
  (`x/exp`) e sem armazém de relações.
- **OpenFGA embutido:** ReBAC mais completo, mas mais dependências, histórico
  de bypass em casos de borda do avaliador e escrita fora da transação da
  aplicação.
- **SpiceDB ou OpenFGA como serviço:** um serviço e um banco a mais (custo
  fixo, premissa 17), modelo só relacional e dependência de fornecedor.

## Consequências

- O motor é nosso, e os bugs também. Verificação obrigatória:

| Verificação | Quando |
|---|---|
| Testes diferenciais contra `cedar-go` no núcleo da linguagem | Cada PR (curto) e noite (amplo) |
| Fuzzing de parser e avaliador, com limites de tamanho e profundidade | Cada PR (curto) e noite (longo) |
| Testes de propriedade: check × lote × listagem coerentes | Cada PR |
| Fail-closed: erro no meio da avaliação nega (teste de abuso) | Cada PR |
| Auditoria externa | Antes de produção |

- `cedar-go` nunca entra no binário: dependência só de teste (design,
  seção 10).
- Avaliação de políticas é área sensível: merge só com aprovação do dono
  (brief 4.28).
- A linguagem precisa de um núcleo comparável com o oráculo
  ([ADR 0015](0015-authz-cedar-core-language.md)).
