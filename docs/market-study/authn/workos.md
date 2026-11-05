# WorkOS (AuthKit, SSO, Directory Sync)

> Escopo: autenticação (AuthKit, SSO, Directory Sync, Admin Portal, Radar). O WorkOS FGA é tratado em documento separado de AuthZ.

| Campo | Valor |
|---|---|
| Categoria | IDaaS B2B "enterprise-ready": APIs de SSO, SCIM, Admin Portal, Audit Logs, Radar, Vault, e o AuthKit (gestão de usuários completa) |
| Site / repositório | https://workos.com · docs https://workos.com/docs · SDKs e AuthKit em https://github.com/workos (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário; SDKs e componentes AuthKit abertos (licença não verificada: GitHub inacessível). **AuthKit grátis até 1 mi MAU**, depois US$ 2.500/mês por milhão. **SSO e Directory Sync cobrados por conexão**: US$ 125/mês (1–15), descendo até US$ 65 (51–100). Audit Log streaming US$ 125/mês por SIEM; Radar US$ 100 por 50 mil checagens; domínio customizado US$ 99/mês; staging grátis |
| Stack | Fechada; não verificada |
| Maturidade | Lançamento público em mar/2020 (Show HN 22607402, 205 pontos); Série B de US$ 80 mi em 2022; comprou a Warrant (FGA) em abr/2024. `@workos-inc/node` ~14 mi downloads/mês, `@workos-inc/authkit-nextjs` ~3,7 mi (npm, out/2026) |
| Acesso | 2026-11-03 |

> Limitações: GitHub e Reddit bloqueados nesta sessão; advisories via OSV.dev; página de segurança lida em HTML.

## Resumo
- Nasceu como "Stripe para recursos enterprise" (SSO/SCIM por API) e virou CIAM completo com o AuthKit.
- Modelo B2B **organization-first**: organizações com domínios verificados, política de domínio, política por organização, papéis por organização.
- **Admin Portal** (link para o TI do cliente configurar SSO/SCIM/domínio sozinho) é o diferencial histórico.
- Preço "grátis para usuários, caro por conexão": atrai startups, pesa quando SSO vira item de planos baratos.
- Incidente de out/2026: invasão de sistemas internos de vendas/marketing (CRM), sem evidência de acesso a produção.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | E-mail+senha, *Magic Auth* (código de 6 dígitos por e-mail; **link mágico depreciado** porque scanners de e-mail corporativo consomem o link), passkeys, social, SSO SAML/OIDC (dezenas de IdPs). LDAP direto e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ◐ | AuthKit hospedado: só TOTP. SMS só pela MFA API avulsa e **só números dos EUA**. Reautenticação (step-up) documentada. Push, WhatsApp, MFA adaptativo: não (Radar desafia com OTP por e-mail) |
| 3 | Ciclo de vida da conta | ● | Verificação de e-mail obrigatória por padrão, convites, *invite-only signup*, *waitlist* com aprovação, JIT por domínio verificado, Directory Sync (SCIM, Google Workspace, Workday etc.), membership `pending/active/inactive` com desativação/reativação, mudança de e-mail. Verificação de idade: não encontrada |
| 4 | Sessões e tokens | ◐ | Access token JWT + refresh token com **rotação e janela de replay de 30 s**; timeout por inatividade; revogação de sessão e de todas as sessões; troca de organização via refresh token. Single Logout só OIDC e "cenários limitados". DPoP: não encontrado |
| 5 | Protocolos | ◐ | SAML e OIDC como SP; AuthKit como servidor de autorização OAuth/OIDC (MCP) com **CIMD** e DCR, *resource indicators* (`aud` por recurso); SCIM de entrada. CIBA, PAR, FAPI, device flow: não encontrados |
| 6 | Máquinas e agentes | ◐ | API keys por usuário e por organização (widgets), M2M/OAuth apps ("Connect"), AuthKit para servidores MCP, "Pipes" (tokens de terceiros para agentes). Identidade de workload: não encontrada |
| 7 | Detecção e defesa | ● | **Radar**: fingerprint de dispositivo, bot detection, credential stuffing, viagem impossível, ações bloquear/desafiar (OTP por e-mail ou SMS no cadastro)/notificar |
| 8 | Privacidade e conformidade | ◐ | SOC 2 Type 2, GDPR/CCPA, BAA HIPAA (enterprise). **Vault** (EKM/BYOK por contexto). Residência de dados/região: não encontrada nas docs nem na página de segurança |
| 9 | Experiência e administração | ● | AuthKit hospedado e APIs para UI própria, **Admin Portal** e Widgets (usuários, sessões, SSO, domínio, API keys, audit log), Audit Logs como produto, eventos/webhooks com reconciliação, impersonação com motivo obrigatório |
| 10 | Operação e migração | ◐ | Guias de migração de Auth0, Clerk, Cognito, Descope, Firebase, Stytch, Supabase, Better Auth; guia para clientes on-prem (um ambiente/API key por instalação). Só SaaS |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Admin Portal**: link de configuração para o TI do cliente (expira em 30 dias pelo dashboard, 5 min via API), escopo restrito a uma organização.
- **Política de domínio vs. política de organização**: a de domínio vale para todo usuário com aquele e-mail; a de organização vale ao entrar nela (exigir SSO ou MFA, útil para convidados sem o domínio).
- **Domínios compartilhados**: várias organizações podem verificar o mesmo domínio (unidades de negócio com IdPs diferentes); descoberta de SSO pergunta qual organização quando há ambiguidade.
- **SSO obrigatório ao configurar**: ao criar a primeira conexão SSO, os outros métodos da organização são desativados.
- **Hierarquia Team → Project → Environment → Application**; e "AuthKit for Platforms" provisiona um ambiente por cliente via Platform API.
- **Vault** com chave por contexto (ex.: por organização) e BYOK no ambiente do cliente.

## O que a comunidade elogia (com evidência)
- Clareza da documentação e rapidez para entregar SSO — agregadores de avaliações (rfp.wiki, everydev.ai; G2 4,5 com só 15 avaliações). Evidência fraca (sites possivelmente gerados por IA).
- Posicionamento "enterprise-ready por API" bem recebido no lançamento (HN 22607402, 30767084).

## Do que a comunidade se queixa (com evidência)
- **Preço por conexão ("SSO tax")**: "WorkOS charges $125/month for SCIM (on top of 125/month for SSO)" (HN 44196896, 2025); avaliações citam custo proibitivo para SSO em planos baratos (rfp.wiki, agregador).
- **Recursos B2C limitados**: MFA por SMS só EUA e fora do AuthKit; link mágico depreciado (docs).
- **Incidente out/2026**: acesso a API interna de vendas/marketing (Salesforce/Snowflake) por ataque atribuído a framework autônomo de pentest com IA; comentário cético: "We don't know how we were hacked" (HN 50345712, notificação de cliente reproduzida pelo autor; CEO respondeu no fio).

## Histórico de segurança

| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVE-2024-29901 / GHSA-35w3-6qhc-474v | 2024 | Média, replay de sessão expirada controlando o cabeçalho `x-workos-session` (authkit-nextjs) | Nunca aceitar estado de sessão vindo de cabeçalho controlável pelo cliente |
| CVE-2024-51752 / CVE-2024-51753 | 2024 | Baixa, refresh tokens logados com `debug` ativo (Next.js e Remix) | Redação de segredos no logger por tipo, inclusive em modo debug |
| CVE-2025-55008 / CVE-2025-55009 | 2025 | Alta, `sealedSession` e access token renderizados no HTML pelo loader (React Router/Remix) | SDKs não devem expor tokens ao cliente por padrão; contrato explícito do que sai do servidor |
| CVE-2025-64762 / GHSA-p8pf-44ff-93gf | 2025 | Alta, respostas autenticadas sem cabeçalhos anti-cache → CDN servia token de sessão a outros | Middleware de autenticação deve forçar `Cache-Control: no-store`/`private` em toda resposta autenticada |
| Incidente CRM (sem CVE) | 2026 | Acesso não autorizado a sistemas internos de GTM | Isolar dados de clientes de ferramentas internas; ataques automatizados por IA aceleram exploração |

## Arquitetura e decisões notáveis
- E-mail é o identificador único do usuário; **vinculação de identidade só após provar posse do e-mail**; domínio verificado implica e-mails verificados para usuários do SSO daquele domínio (e só deles).
- Normalização de e-mail (trim + minúsculas) documentada.
- Admin Portal e Widgets usam sessões escopadas à organização.
- Ambientes isolados (usuários, organizações, conexões e chaves por ambiente); algumas configurações vivem no projeto.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Modelo**: ambiente (equivale a tenant isolado) → usuários únicos por e-mail → organizações com memberships e papel por membership; suporta "um usuário, várias organizações" e "uma organização por usuário".
- Papéis no nível do ambiente + **papéis customizados por organização**; atribuição de papel pelo IdP (grupos SSO/SCIM).
- Por organização: domínios, conexões SSO/SCIM, políticas de domínio e de organização (SSO/MFA), papéis, metadados, API keys.
- Isolamento total por cliente: um ambiente por cliente (Platforms ou on-prem).
- Extensão: eventos/webhooks, JWT templates, metadados, feature flags. Hooks síncronos no login: não encontrados.
- Soberania: não encontrada (região única presumida; não verificado).

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não**. Só aprovação individual de *waitlist* (docs AuthKit Waitlist).
- Autorização sensível ao estado / transições: **não** no escopo AuthN (FGA avaliado em outro documento).

## Implicações para o projeto
1. **Adotar** política de domínio + política de organização como dois níveis distintos; exigir SSO automaticamente quando a organização configura um IdP (com alerta, não bloqueio, se quiser manter outros métodos).
2. **Adotar** um "Admin Portal" de autoatendimento por link temporário escopado à organização para SSO, SCIM e verificação de domínio.
3. **Adotar** domínios compartilhados entre organizações com descoberta desambiguada.
4. **Adotar** rotação de refresh token com janela de replay curta e idempotente para concorrência.
5. **Evitar** link mágico como padrão em B2B (scanners consomem o link); preferir código.
6. **Oportunidade**: sem cobrança por conexão SSO/SCIM (queixa recorrente) e com residência de dados.
7. **Adotar** cabeçalhos anti-cache e redação de tokens em logs nos SDKs, com testes (lições dos CVEs dos SDKs).

## Fontes
- https://workos.com/docs/llms.txt — 2026-11-03 — oficial
- https://workos.com/docs/authkit/users-organizations.md, /modeling-your-app.md, /organization-policies.md, /platforms.md, /projects.md — 2026-11-03 — oficial
- https://workos.com/docs/authkit/sessions.md, /session-resilience.md, /mfa.md, /mcp.md, /impersonation.md, /radar.md, /identity-linking.md — 2026-11-03 — oficial
- https://workos.com/docs/admin-portal.md, /sso/single-logout.md, /sso/saml-security.md, /rbac/custom-roles.md, /vault.md, /on-prem-deployment.md — 2026-11-03 — oficial
- https://workos.com/pricing.md — 2026-11-03 — oficial
- https://workos.com/security — 2026-11-03 — oficial
- https://api.osv.dev/v1/vulns/GHSA-35w3-6qhc-474v, GHSA-5wmg-9cvh-qw25, GHSA-v2qh-f584-6hj8, GHSA-vqvc-9q8x-vmq6, GHSA-v3gr-w9gf-23cx, GHSA-p8pf-44ff-93gf — 2026-11-03 — base pública
- https://news.ycombinator.com/item?id=50345712 — 2026-11-03 — comunidade (notificação reproduzida por cliente; não confirmada no site da WorkOS na data)
- https://news.ycombinator.com/item?id=22607402, 30767084, 44196896 — 2026-11-03 — comunidade (30767084 é lançamento de concorrente, BoxyHQ)
- https://www.rfp.wiki/vendors/workos — 2026-11-03 — agregador, evidência fraca
- https://api.npmjs.org/downloads/point/last-month/@workos-inc/node — 2026-11-03 — npm
