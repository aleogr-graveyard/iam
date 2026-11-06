# ADR 0008: versão única dos componentes de produto

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.2 decidiu o versionamento híbrido (componentes de produto com uma
versão de produto; SDKs com versão própria pela versão da API) e deixou em
aberto (pergunta 7) se haveria uma versão única da plataforma ou uma por
produto (AuthN e AuthZ). AuthN e AuthZ saem no mesmo binário (`cmd/iam`,
monólito modular).

## Decisão

1. **Versão única** para os componentes de produto.
2. **SDKs com versão própria**, pela versão da API (v1, v2), não pela do
   servidor.

## Alternativas

- **Versão por produto (AuthN e AuthZ):** releases independentes, mas exigiria
  GoReleaser Pro, um contorno frágil ou um pipeline mais complexo (brief 4.2),
  e duas versões de produto conviveriam no mesmo binário.

## Consequências

- Uma release cobre os dois produtos; uma correção só no AuthZ também gera nova
  versão do conjunto.
- Módulos Go publicáveis recebem tags com o prefixo do caminho, como o Go exige
  (por exemplo, `authz/core/v1.4.0`).
- O pipeline de release (brief 4.18) fica mais simples: uma versão, um
  changelog de produto.
- A política de suporte declarada é por versão de produto
  ([ADR 0028](0028-cyber-resilience-act-from-milestone-0.md)).
- Separar as versões no futuro exige nova decisão.
