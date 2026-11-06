# ADR 0028: obrigações do Cyber Resilience Act desde o marco 0

- Status: aprovado (2026-11-06)

## Contexto

O Cyber Resilience Act (Regulamento (UE) 2024/2847) tem obrigações de
notificação desde 11/09/2026 e aplicação plena em 11/12/2027 (brief, seção 7).
Sistemas de gestão de identidade e de acesso estão no Anexo III, Classe I, e o
Regulamento de Execução (UE) 2025/2392 inclui autenticação **e** autorização
(`docs/market-study/standards/security-references.md`). Open source fora de
contexto comercial fica fora do escopo; monetizar torna o projeto fabricante.
O brief 4.2 e 4.30 já pedem política de suporte declarada e SBOM.

## Decisão

Desde o marco 0:

| Item | Forma |
|---|---|
| `SECURITY.md` | Com o canal privado de reporte de vulnerabilidades do GitHub |
| Processo de tratamento de vulnerabilidades | Documentado |
| SBOM | Em cada build |
| Política de suporte | Declarada |

Enquadramento:

- O iam é **Anexo III, Classe I** (AuthN e AuthZ).
- Com monetização, o projeto passa a ser **fabricante**.
- As obrigações de notificação já valem desde **11/09/2026**.

## Alternativas

- **Adiar até monetizar ou até 11/12/2027:** o processo e os artefatos custam
  pouco no início e muito para recuperar depois; a notificação já está em
  vigor.

## Consequências

- Entrega do marco 0: `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, política de
  suporte e SBOM no build (design, seção 12).
- A política de suporte é por versão de produto
  ([ADR 0008](0008-single-product-version.md)).
- Antes de monetizar ([ADR 0009](0009-apache-2-license.md)): validar com
  jurídico a avaliação de conformidade da Classe I.
- O mínimo de dependências (premissa 24) e o `govulncheck` em cada PR reduzem
  a superfície a tratar.
