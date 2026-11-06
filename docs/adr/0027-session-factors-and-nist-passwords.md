# ADR 0027: sessão com fatores carimbados e senhas pelo NIST SP 800-63-4

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.5 propõe que a sessão central registre quais fatores foram cumpridos
e quando, para que um app com exigência maior peça só o fator que falta. O
Spring Security 7 adota esse modelo: cada fator é uma autoridade com carimbo de
tempo, e a política exige fatores e frescor. O NIST SP 800-63-4 (final em
julho de 2025) revisou as regras de senha
(`docs/market-study/standards/security-references.md`). O estudo registra o
bcrypt truncando em 72 bytes.

## Decisão

**Sessão central**

1. Registra **cada fator cumprido com horário**.
2. A **política exige fatores e frescor**; o AuthN pede só o que falta.

**Senhas (NIST SP 800-63-4)**

| Regra | Valor |
|---|---|
| Mínimo como fator único | 15 caracteres |
| Mínimo como parte de MFA | 8 caracteres |
| Máximo aceito | pelo menos 64 caracteres |
| Regras de composição | nenhuma |
| Troca periódica | nenhuma |
| Lista de bloqueio e senhas vazadas | checagem obrigatória |
| Truncamento | nenhum |

## Alternativas

- **Sessão com um único estado "autenticado"** (ou um nível só): não diz quais
  fatores foram cumpridos nem quando, e obriga a repetir o login inteiro no
  step-up.
- **Regras de composição e troca periódica:** o NIST SP 800-63-4 as proíbe.

## Consequências

- O `acr` e o `amr` dos tokens saem dos fatores registrados; a obrigação de
  step-up do AuthZ liga ao `acr` ([ADR 0020](0020-authz-dynamic-access.md)).
- Hash com Argon2id (brief 4.10); nada de algoritmo que trunque a senha.
- Senhas vazadas checadas por lista local no marco 1 (design, seção 6).
- Sessão e fluxo de login são máquinas de estados validadas no servidor:
  impossível chegar a "autenticado" sem os fatores exigidos (teste baseado em
  modelo).
- Sessões e tokens são área sensível (brief 4.28).
