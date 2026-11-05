# Autorização embutida em produtos de identidade (comparativo)

Keycloak Authorization Services · Zitadel · Auth0 (RBAC + Actions) · Microsoft Entra ID (app roles, grupos, PIM)

| Campo | Valor |
|---|---|
| Categoria | Autorização oferecida "de brinde" pelo IdP, ao lado de AuthN |
| Site / repositório | Keycloak: https://www.keycloak.org/docs/latest/authorization_services/ · Zitadel: https://zitadel.com/docs · Auth0: https://auth0.com/docs/manage-users/access-control · Entra: https://learn.microsoft.com/entra/identity-platform/how-to-web-app-role-based-access-control |
| Licença e modelo de negócio | Keycloak: Apache-2.0 (Red Hat build pago). Zitadel: **AGPL-3.0-only** no servidor; `proto/` e docs Apache-2.0; login e clientes MIT (LICENSING.md); nuvem e licença comercial pagas. Auth0: SaaS proprietário (Okta); Auth0 FGA é produto separado. Entra: SaaS proprietário; PIM e entitlement management exigem licenças P2/ID Governance |
| Stack | Keycloak: Java/Quarkus, bancos relacionais. Zitadel: Go, PostgreSQL, event sourcing. Auth0/Entra: não publicados |
| Maturidade | Todos maduros e muito adotados como IdP; a parte de autorização é secundária em todos |
| Acesso | 2026-11-03 |

## Resumo
- **Keycloak** é o único com um motor de políticas de verdade embutido (recursos, escopos, políticas, permissões, UMA 2.0), mas é acoplado ao realm/cliente e com histórico recente de falhas de autorização no próprio motor.
- **Zitadel** oferece só RBAC por projeto ("role assignments", antes "user grants") com concessão de projeto entre organizações; para autorização fina, a própria Zitadel indica lógica na aplicação ou ferramentas de terceiros, e usuários combinam com OpenFGA.
- **Auth0** tem RBAC por API (permissões no token) e Actions (código no login); autorização fina foi para outro produto (Auth0 FGA/OpenFGA). Suporta RAR com tela de consentimento e push no Guardian.
- **Entra** tem app roles e grupos no token (com limite de 200 grupos e "overage"), mais PIM (JIT com aprovação) e entitlement management (pacotes de acesso com até 3 estágios).
- Padrão comum: **papéis amplos no token**; nenhum oferece checagem fina por requisição, listagem, explicabilidade nem AuthZEN (não encontrado em nenhum dos quatro).

## Funcionalidades por domínio

| Domínio (5.2) | Keycloak AS | Zitadel | Auth0 | Entra | Notas |
|---|---|---|---|---|---|
| 1. Modelos | ◐ | ◐ | ◐ | ◐ | KC: RBAC, UBAC, grupos, tempo, regex, client scope, JS, agregadas (ABAC limitado); sem ReBAC. Zitadel/Auth0/Entra: RBAC (+ organizações/tenants) |
| 2. Decisões | ◐ | ○ | ○ | ○ | KC: avaliação de permissões e RPT com todas as permissões do usuário (lista "o que posso"); demais: só claims no token |
| 3. Políticas | ◐ | ○ | ◐ | ○ | KC: UI + Policy Evaluation Tool (simulação); políticas JS só via JAR implantado; sem versionamento/testes. Auth0: Actions (código JS versionado no tenant) |
| 4. Dados | ◐ | ○ | ○ | ○ | KC: recursos registrados via Protection API (sincronização manual pela aplicação) |
| 5. Acesso dinâmico | ◐ | ○ | ○ | ● | KC: UMA — dono do recurso aprova/nega pedidos (1 aprovador); política de tempo. Entra: PIM (JIT, aprovação, 24 h) e pacotes de acesso com expiração |
| 6. Estados e transições | ○ | ○ | ○ | ○ | Nenhum |
| 7. Padrões | ◐ | ◐ | ◐ | ◐ | KC: UMA 2.0. Auth0: RAR (`authorization_details`). Todos: escopos OAuth. AuthZEN: não encontrado em nenhum |
| 8. Máquinas e agentes | ◐ | ◐ | ◐ | ◐ | Service accounts / M2M em todos; Entra Agent ID com pacotes de acesso para agentes; nada de autorização por ferramenta MCP no IdP |
| 9. Desempenho e resiliência | ◐ | ● | ● | ● | Claims no token = decisão local, sem latência; KC AS exige chamada ao servidor (policy enforcer com cache) |
| 10. Auditoria | ◐ | ◐ | ◐ | ● | Eventos administrativos e de login; log de decisões de autorização fina: não (exceto KC parcial). Entra: logs de auditoria do PIM |
| 11. Experiência e administração | ● | ● | ● | ● | Todos com console; é o ponto forte |

## Diferenciais / exclusivos
- **Keycloak:** estratégias de decisão Unanimous/Affirmative/**Consensus** para combinar políticas (consensus = mais positivas que negativas — combinação de políticas, não aprovação humana); fluxo UMA de compartilhamento usuário-a-usuário com aprovação do dono; Policy Evaluation Tool.
- **Zitadel:** concessão de projeto a outra organização (o "cliente do cliente" recebe papéis do projeto) — modelo B2B nativo e simples; claims de papéis por projeto (`urn:zitadel:iam:org:project:{projectId}:roles`).
- **Auth0:** RAR com renderização em tela de consentimento e no app Guardian (aprovação por push de um pedido detalhado — 1 pessoa, o próprio usuário).
- **Entra:** governança integrada (PIM, access packages, revisões de acesso) — mais rica em aprovação que em decisão.

## O que a comunidade elogia (com evidência)
- Keycloak: gratuito, completo, UI; a documentação de AS cobre ABAC/RBAC/UMA num só lugar (guia oficial).
- Zitadel: multi-tenancy B2B simples com organizações e grants (doc oficial "Retrieve user roles").
- Auth0/Entra: integração imediata — papéis já chegam no token sem serviço extra.

## Do que a comunidade se queixa (com evidência)
- **Keycloak:** lentidão da Admin API com fine-grained admin permissions (RHBK-1852); FGAP v1 depreciado, migração obrigatória para v2 (release notes RHBK 26.x); adaptadores próprios descontinuados e "as alternativas não suportam Authorization Services" (blog Keycloak, 2023) — quem usava policy enforcer ficou sem caminho oficial em várias linguagens; opções de claims do token endpoint de AS depreciadas para remoção no Keycloak 27.
- **Zitadel:** autorização fina inexistente; usuário relata que rodar OpenFGA ao lado do Zitadel "adiciona overhead operacional" (questions.zitadel.com, mai/2025); a própria Zitadel abriu pesquisa sobre o tema.
- **Auth0:** permissões no token aumentam o tamanho do token (doc oficial); limites de 50 papéis por usuário e 1000 permissões por papel/usuário; não há forma confiável de ler permissões numa Action sem a Management API, que tem rate limit baixo (fórum Auth0, relato de usuário).
- **Entra:** **group overage** — acima de 200 grupos (JWT) ou 150 (SAML) o claim some e a aplicação precisa chamar o Graph; implicit flow com limite ainda menor. PIM sem aprovação multietapas e com janela de 24 h não configurável.

## Histórico de segurança
| ID | Ano | Produto | Classe | Lição |
|---|---|---|---|---|
| CVE-2025-14778 | 2025 | Keycloak (UMA) | Avaliação checava só o **primeiro recurso** da requisição | Em lote, avaliar e negar item a item; teste de propriedade "um item negado nunca vira permitido" |
| CVE-2026-4636 | 2026 | Keycloak (UMA) | Injeção de IDs de recursos alheios ao criar política (`uma_protection`) → RPT para recursos de vítimas | Toda referência a recurso em escrita de política deve passar por checagem de posse/tenant |
| CVE-2026-15573 | 2026 | Keycloak (PathMatcher) | Casamento lexical sem normalização de URI (matrix param, barra final) cai em regra ampla `/*` | Normalizar antes de casar; regras por caminho falham para regra mais restrita, nunca para mais ampla |
| CVE-2026-18572 | 2026 | Keycloak (política de tempo) | Claim de tempo forjada pelo cliente sobrescreve o relógio do servidor | Atributos de ambiente (hora, IP) vêm do servidor; contexto do cliente nunca sobrescreve atributos confiáveis |
| CVE-2025-27507 | 2025 | Zitadel | IDOR na Admin API (CWE-639) | Escopo por organização/tenant aplicado em toda consulta administrativa |
| GHSA-wr6r-59xg-4pj2 | 2026 (data não confirmada) | Zitadel | Dados do plano de gestão de outra organização via `project_id`/`grant_id`/`app_id` | Idem: isolamento de tenant em IDs secundários, não só no principal |
| CVE-2022-36051 | 2022 | Zitadel (Actions) | Autorização quebrada em código JS executado no login | Extensões com código são superfície de ataque; preferir configuração declarativa |

## Arquitetura e decisões notáveis
- Keycloak: AS é por **cliente** (resource server) dentro do realm; políticas avaliadas no servidor; RPT carrega permissões concedidas — token como cache de decisão.
- Zitadel: event sourcing; autorização = projeção de "role assignments" em claims.
- Auth0: Authorization Core (RBAC) + Actions que podem sobrescrever escopos definidos pelo RBAC (doc oficial) — duas fontes de verdade.
- Entra: app roles definidos no manifesto da aplicação, atribuídos a usuários/grupos; PIM eleva papéis temporariamente.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Keycloak: realms como tenants; SPI Java para políticas customizadas; Postgres/MySQL/MariaDB/Oracle/MSSQL; self-hosted.
- Zitadel: instâncias → organizações → projetos; Actions v2 (webhooks); PostgreSQL; self-hosted ou nuvem (regiões).
- Auth0: tenants e Organizations (limites RBAC passam a valer por organização); Actions em JS; SaaS (regiões, Private Cloud).
- Entra: tenant Microsoft; extensões via Graph e custom claims providers; SaaS.

## (Somente AuthZ) Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Modelos:** RBAC em todos; ABAC limitado só no Keycloak.
- **Linguagem:** nenhuma linguagem declarativa de política (Keycloak: UI + JS).
- **Consistência:** papéis no token ficam desatualizados até o token expirar (problema clássico; nenhum resolve).
- **Listagem:** só Keycloak (RPT com todas as permissões); nenhum lista "quem acessa o recurso X".
- **Sincronização:** Keycloak exige registrar recursos no IdP; demais não modelam recursos.
- **Explicabilidade/testes:** só a Policy Evaluation Tool do Keycloak.
- **Avaliação local:** claims no token (todos).
- **AuthZEN:** não encontrado em nenhum dos quatro (documentação oficial consultada; busca adicional não foi possível por limite de buscas da sessão).

## Verificação de diferenciais candidatos do projeto
- **Aprovação M de N de primeira classe?** **Não** em nenhum dos quatro. Keycloak UMA: aprovação pelo dono (1 pessoa). Entra PIM: "requests are resolved by the first approver who approves or denies"; "multi-step approvals aren't currently supported". Entitlement management: até 3 estágios sequenciais, mas "apenas um aprovador de cada estágio precisa aprovar". Auth0/Zitadel: não há fluxo de aprovação.
- **Autorização sensível ao estado / transições declarativas?** **Não** em nenhum. No Keycloak, estado poderia entrar só via política JS ou claim.

## Implicações para o projeto
1. **Adotar:** "papéis amplos no token, permissões finas no PDP" — confirmado como prática de mercado e pelas queixas (overage do Entra, tamanho de token do Auth0).
2. **Adotar:** concessão de projeto/papéis a organizações de clientes (modelo Zitadel) para o nosso caso B2B2C.
3. **Evitar:** duas fontes de verdade (Auth0: Actions sobrescrevem RBAC). Uma só política; extensões não podem contornar o PDP.
4. **Evitar:** atributos de ambiente vindos do cliente (CVE-2026-18572) e casamento de caminhos sem normalização (CVE-2026-15573); entram no modelo de ameaças.
5. **Oportunidade:** AuthN + AuthZ irmãos com contrato AuthZEN — nenhum IdP do grupo oferece PDP fino, listagem, explicação ou AuthZEN; todos empurram o cliente para um segundo produto.
6. **Oportunidade:** aprovação com quórum real (Entra só tem "primeiro que responder"); vale para PIM da própria plataforma (seção 4.23).

## Fontes
- https://www.keycloak.org/docs/latest/authorization_services/index.html — 2026-11-03 — oficial
- https://www.keycloak.org/2023/03/adapter-deprecation-update — 2026-11-03 — blog oficial
- https://docs.redhat.com/en/documentation/red_hat_build_of_keycloak/26.2/html/release_notes/deprecated_features — 2026-11-03 — oficial Red Hat
- https://www.keycloak.org/docs/26.7.3/upgrading/ — 2026-11-03 — oficial
- https://issues.redhat.com/browse/RHBK-1852 — 2026-11-03 — issue tracker
- https://startwithidentity.com/cves/cve-2025-14778/ — 2026-11-03 — resumo de terceiro
- https://advisories.gitlab.com/maven/org.keycloak/keycloak-services/CVE-2026-4636/ — 2026-11-03 — base de advisories
- https://bugzilla.redhat.com/show_bug.cgi?id=2499593 — 2026-11-03 — Red Hat (CVE-2026-15573)
- https://bugzilla.redhat.com/show_bug.cgi?id=2509763 — 2026-11-03 — Red Hat (CVE-2026-18572, status NEW)
- https://raw.githubusercontent.com/zitadel/zitadel/main/LICENSING.md — 2026-11-03 — oficial
- https://zitadel.com/docs/guides/integrate/retrieve-user-roles — 2026-11-03 — oficial
- https://zitadel.com/blog/fine-grained-authorization — 2026-11-03 — blog do fornecedor
- https://questions.zitadel.com/m/1368861057669533736 — 2026-11-03 — fórum da comunidade
- https://db.gcve.eu/vuln/cve-2025-27507 — 2026-11-03 — base CVE
- https://osv.dev/vulnerability/GHSA-c8fj-4pm8-mp2c — 2026-11-03 — OSV
- https://auth0.com/docs/manage-users/access-control/configure-core-rbac/enable-role-based-access-control-for-apis — 2026-11-03 — oficial
- https://auth0.com/docs/troubleshoot/customer-support/operational-policies/entity-limit-policy — 2026-11-03 — oficial
- https://auth0.com/docs/get-started/apis/configure-rich-authorization-requests.md — 2026-11-03 — oficial
- https://community.auth0.com/t/action-login-flow-not-able-to-enrich-the-access-token-with-the-users-permission/116930 — 2026-11-03 — fórum (relato individual)
- https://www.cerbos.dev/blog/you-have-auth0-what-authorization-capabilities-do-you-still-need — 2026-11-03 — concorrente (viés)
- https://learn.microsoft.com/entra/identity-platform/how-to-web-app-role-based-access-control — 2026-11-03 — oficial
- https://learn.microsoft.com/azure/active-directory/hybrid/how-to-connect-fed-group-claims — 2026-11-03 — oficial (limite de grupos)
- https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-approval-workflow — 2026-11-03 — oficial
- https://docs.azure.cn/en-us/entra/id-governance/entitlement-management-access-package-approval-policy — 2026-11-03 — oficial (espelho China)
