# ADR 0009: licença Apache-2.0

- Status: aprovado (2026-11-06)
- Escopo: licença do código. Monetização continua em aberto (design,
  seção 14).

## Contexto

O brief (seção 8, pergunta 9) deixou a licença em aberto entre Apache-2.0,
MIT, AGPL-3.0 e source-available (BSL, FSL). A premissa 11 pede open source por
concepção, com possibilidade de monetização. O repositório é público. O estudo
de mercado registra volatilidade de licenças como queixa recorrente
(`docs/market-study/summary.md`, seção 3).

## Decisão

1. **Apache-2.0 em todo o repositório.**
2. **CLA** (não apenas DCO) se houver contribuições externas.
3. Modelo de monetização ainda em aberto.

## Alternativas

- **AGPL-3.0 + licença comercial** (modelo Zitadel): protege contra concorrência
  hospedada, mas reduz a adoção.
- **Source-available (FSL, BSL):** não é open source; contraria a premissa 11.
- **MIT:** permissiva, mas sem concessão explícita de patentes.

## Consequências

- Qualquer um pode hospedar o produto comercialmente.
- Recursos pagos, se existirem, entram por pontos de extensão, sem bifurcar o
  código (design, seção 14).
- Dependências precisam de licença compatível; verificação de licenças em cada
  PR (design, seção 11).
- O CLA preserva a possibilidade de licenciamento comercial.
- Monetizar torna o projeto fabricante perante o Cyber Resilience Act
  ([ADR 0028](0028-cyber-resilience-act-from-milestone-0.md)).
