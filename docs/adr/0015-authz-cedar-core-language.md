# ADR 0015: linguagem de políticas com núcleo compatível com Cedar

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.29 pergunta se a linguagem de políticas é própria ou existente
(Cedar, Rego, a linguagem de modelagem do OpenFGA). O Cedar foi desenhado para
ser analisável: `forbid` vence, o padrão é negar, sem laços, regex nem ponto
flutuante, com modelo formal em Lean (`docs/market-study/authz/cedar.md`). Os
dois advisories do Cedar em 2026 foram divergência entre JSON e texto e
exaustão de memória no parser. O motor é próprio, com `cedar-go` como oráculo
([ADR 0014](0014-authz-own-engine-with-oracle.md)).

## Decisão

1. **Núcleo compatível com Cedar:** `permit`/`forbid`,
   `principal`/`action`/`resource`, `when`/`unless`, `in`; sem laços, regex nem
   ponto flutuante.
2. **Extensões explícitas** (estados e transições, relações com prazo),
   **traduzíveis para o núcleo** nos testes.
3. **Representação JSON** com a mesma validação do texto.

## Alternativas

- **Linguagem própria:** sem oráculo para testes diferenciais e com uma
  linguagem nova a aprender.
- **Rego (OPA):** Datalog genérico, mais difícil de analisar e mais lento nas
  medições publicadas.
- **Só dados (YAML/JSON):** condições sobre atributos e contexto virariam uma
  linguagem improvisada dentro dos dados.

## Consequências

- Gramática, esquema JSON e corpus de testes formam o contrato C10 (design,
  seção 8).
- Testes diferenciais entre texto e JSON, além de contra o oráculo.
- Como as extensões se traduzem para o núcleo, o oráculo cobre também
  políticas que as usam.
- Fuzzing do parser com limites de tamanho e profundidade para políticas
  vindas de tenants.
- Análise formal de políticas fica possível depois
  ([ADR 0019](0019-authz-policy-lifecycle.md)).
