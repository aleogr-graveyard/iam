# Matriz de funcionalidades — AuthZ

Acesso: 2026-11-03 a 2026-11-05. Legenda: ● completo, ◐ parcial, ○ ausente, ? não verificado.
Avaliação qualitativa de cada agente de pesquisa; detalhes e fontes no documento de cada player.

Domínios (brief, seção 5.2):
1. Modelos · 2. Decisões (check, lote, listagem, explicabilidade) · 3. Políticas ·
4. Dados (relações, atributos, sincronização, consistência) · 5. Acesso dinâmico (temporário, JIT,
aprovações, M de N, delegação, obrigações) · 6. Estados e transições · 7. Padrões ·
8. Máquinas e agentes · 9. Desempenho e resiliência · 10. Auditoria · 11. Experiência e administração

| Player | Categoria | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| [OpenFGA](authz/openfga.md) | Zanzibar | ◐ | ◐ | ◐ | ◐ | ◐ | ○ | ◐ | ◐ | ◐ | ◐ | ◐ |
| [SpiceDB](authz/spicedb.md) | Zanzibar | ◐ | ● | ◐ | ● | ◐ | ○ | ○ | ◐ | ◐ | ◐ | ◐ |
| [Ory Keto](authz/ory-keto.md) | Zanzibar | ◐ | ◐ | ◐ | ◐ | ○ | ○ | ○ | ◐ | ◐ | ? | ◐ |
| [Permify](authz/permify.md) | Zanzibar | ◐ | ◐ | ◐ | ● | ◐ | ○ | ○ | ◐ | ◐ | ? | ◐ |
| [OPA](authz/opa.md) | Motor de políticas | ◐ | ◐ | ● | ◐ | ○ | ◐ | ◐ | ◐ | ● | ● | ◐ |
| [Cedar](authz/cedar.md) | Motor de políticas | ● | ◐ | ● | ◐ | ○ | ◐ | ◐ | ◐ | ● | ○ | ◐ |
| [Cerbos](authz/cerbos.md) | Motor de políticas | ◐ | ● | ● | ◐ | ○ | ◐ | ● | ◐ | ● | ● | ◐ |
| [Casbin](authz/casbin.md) | Motor de políticas | ◐ | ◐ | ◐ | ◐ | ○ | ◐ | ○ | ○ | ◐ | ◐ | ◐ |
| [Permit.io](authz/permit-io.md) | Plataforma | ● | ◐ | ● | ◐ | ◐ | ○ | ◐ | ◐ | ● | ◐ | ● |
| [Oso](authz/oso.md) | Plataforma | ● | ● | ◐ | ● | ◐ | ○ | ○ | ◐ | ◐ | ◐ | ◐ |
| [Aserto/Topaz](authz/aserto-topaz.md) | Plataforma | ● | ● | ◐ | ◐ | ○ | ○ | ● | ◐ | ● | ◐ | ◐ |
| [Amazon Verified Permissions](authz/amazon-verified-permissions.md) | Plataforma | ◐ | ◐ | ◐ | ○ | ◐ | ○ | ◐ | ◐ | ◐ | ◐ | ◐ |
| [WorkOS FGA](authz/workos-fga.md) | Plataforma | ◐ | ◐ | ○ | ◐ | ○ | ○ | ○ | ◐ | ◐ | ◐ | ◐ |
| [PlainID](authz/plainid.md) | PBAC corporativo | ◐ | ● | ◐ | ● | ◐ | ○ | ◐ | ● | ◐ | ● | ● |
| [Axiomatics](authz/axiomatics.md) | PBAC corporativo | ◐ | ● | ● | ◐ | ◐ | ○ | ● | ◐ | ◐ | ● | ◐ |
| [Keycloak Authz Services](authz/idp-embedded-authz.md) | Embutido em IdP | ◐ | ◐ | ◐ | ◐ | ◐ | ○ | ◐ | ◐ | ◐ | ◐ | ● |
| [Zitadel](authz/idp-embedded-authz.md) | Embutido em IdP | ◐ | ○ | ○ | ○ | ○ | ○ | ◐ | ◐ | ● | ◐ | ● |
| [Auth0 RBAC/Actions](authz/idp-embedded-authz.md) | Embutido em IdP | ◐ | ○ | ◐ | ○ | ○ | ○ | ◐ | ◐ | ● | ◐ | ● |
| [Entra app roles](authz/idp-embedded-authz.md) | Embutido em IdP | ◐ | ○ | ○ | ○ | ● | ○ | ◐ | ◐ | ● | ● | ● |
| [Teleport](authz/approvals-pam.md) | Aprovações / PAM | ◐ | ○ | ◐ | ○ | ● | ○ | ○ | ◐ | ? | ● | ● |
| [Vault control groups](authz/approvals-pam.md) | Aprovações / PAM | ◐ | ○ | ◐ | ○ | ● | ○ | ○ | ◐ | ? | ● | ◐ |
| [GCP PAM](authz/approvals-pam.md) | Aprovações / PAM | ◐ | ○ | ◐ | ○ | ● | ○ | ○ | ◐ | ? | ● | ● |
| [AWS MPA/TEAM](authz/approvals-pam.md) | Aprovações / PAM | ○ | ○ | ○ | ○ | ◐ | ○ | ○ | ○ | ? | ● | ◐ |
| [Entra PIM](authz/approvals-pam.md) | Aprovações / PAM | ◐ | ○ | ◐ | ○ | ● | ○ | ○ | ◐ | ? | ● | ● |

## Leitura
- **Domínio 6 (estados e transições): ninguém tem ●.** Onde há ◐, é estado como atributo numa condição.
- **Domínio 5:** quórum M de N existe só em PAM/nuvem (coluna ● nas aprovações), nunca em motores ou
  plataformas de autorização de aplicações.
- **Nenhum player tem ● em todos os domínios 1–4:** quem é forte em relações (Zanzibar) é fraco em
  políticas/ABAC e vice-versa.
- Os critérios de ● variam entre agentes; use a matriz para padrões, não para ranking fino.
