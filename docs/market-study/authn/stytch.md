# Stytch

| Campo | Valor |
|---|---|
| Categoria | IDaaS API-first com dois produtos separados: **Consumer auth** e **Multi-tenant (B2B) auth**; além de antifraude (Device Fingerprinting) e Connected Apps (servidor OAuth/OIDC para apps e agentes) |
| Site / repositório | https://stytch.com · docs https://stytch.com/docs · SDKs em https://github.com/stytchauth (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário; SDKs abertos (licença não verificada). **Subsidiária da Twilio** desde nov/2025 (anúncio 30/out/2025). Pay-as-you-go grátis até 10 mil MAU ("membros e agentes"), 5 conexões SSO **ou** SCIM, 1.000 tokens M2M, 10 mil fingerprints; conexão extra US$ 125; remoção de marca US$ 99; Enterprise sob consulta. SMS repassado por país |
| Stack | Fechada; não verificada |
| Maturidade | Fundada em 2020; avaliação de US$ 1 bi em nov/2021 (HN 29274791). Pacote npm `stytch` ~1,3 mi downloads/mês, `@stytch/vanilla-js` ~0,6 mi (npm, out/2026) |
| Acesso | 2026-11-03 |

> Limitações: GitHub e Reddit bloqueados; advisories via OSV.dev; incidente de 2022 avaliado por fio do HN.

## Resumo
- B2B com modelo **organization-first**: o `Member` é uma conta **dentro** de uma Organization; o e-mail liga Members entre organizações, mas cada org tem seu próprio Member.
- Login por **Discovery** (autentica primeiro, escolhe organização depois) ou login **específico da organização**; *intermediate session* entre etapas (descoberta, MFA).
- Forte em antifraude: **Device Fingerprinting** com veredictos e regras.
- Só EUA (sem residência de dados); a doc argumenta que dados de autenticação costumam ficar fora de exigências regionais.
- Futuro sob Twilio incerto segundo concorrentes (fontes enviesadas).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | Link mágico por e-mail (com PKCE), OTP por e-mail/SMS/**WhatsApp** (consumer), senha (cross-org ou por organização), passkeys/WebAuthn (consumer), OAuth social, SSO SAML/OIDC por organização, carteiras cripto, biometria mobile, *trusted auth tokens* (IdP externo/fator customizado). LDAP e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ◐ | B2B: SMS OTP e TOTP, códigos de recuperação, política de MFA por organização, step-up via intermediate session. Consumer: WebAuthn também. Push: não encontrado. Adaptativo: via DFP (*remembered device*, desafios) |
| 3 | Ciclo de vida da conta | ● | Convites, JIT por domínio de e-mail e por conexão SSO (configurável por organização), SCIM com mapeamento automático de papéis, verificação de e-mail, exclusão. Verificação de idade: não encontrada |
| 4 | Sessões e tokens | ◐ | `session_token` (opaco, revogável) e `session_jwt` (validação local) intercambiáveis; extensão, expiração, revogação, troca de sessão entre organizações, *custom claims*. DPoP, back-channel logout: não encontrados |
| 5 | Protocolos | ◐ | SAML/OIDC como SP, SCIM, **Connected Apps** como servidor OAuth 2.0/OIDC (registro dinâmico para MCP). CIBA, PAR, FAPI, device flow: não encontrados |
| 6 | Máquinas e agentes | ◐ | M2M (client credentials, rotação de segredos, importação), Connected Apps para agentes e servidores MCP com consentimento granular e visibilidade por organização |
| 7 | Detecção e defesa | ● | DFP: fingerprint, veredictos allow/block/challenge, regras, captcha invisível, bloqueio por país, abuso de trial, credential stuffing |
| 8 | Privacidade e conformidade | ◐ | DPF EU/UK/Suíça; página sobre DFP e privacidade. **Só servidores nos EUA** (doc oficial). SOC 2: alegado em 2022 (HN), não reconfirmado |
| 9 | Experiência e administração | ● | UI pré-pronta e headless, **Admin Portal** (membros, configurações da org, SSO, SCIM), temas, impersonação (só pela equipe do fornecedor, não por admins do cliente) |
| 10 | Operação e migração | ◐ | Só SaaS; projetos com ambientes; guias de migração (não detalhados aqui) |
| 11 | Identidade descentralizada | ○ | Não encontrado (carteiras cripto não são VC/EUDI) |

## Diferenciais / exclusivos
- **Member por organização**: identidade, métodos, MFA e metadados escopados à org; o CEO defende que concorrentes são "user-first" e por isso as políticas por organização ficam desajeitadas (HN 39470158, fornecedor).
- **SSO não transferível entre organizações**: uma conexão SSO só autentica na org dona; motivo documentado: IdP corporativo pode atestar qualquer e-mail, e se comprometido daria acesso a outras organizações.
- **Discovery login** + intermediate session tokens como estado explícito entre etapas.
- **Senhas cross-org vs. org-scoped** (configuração de projeto, imutável após haver senhas).
- **Regras de configuração da organização** documentadas formalmente (ex.: `RESTRICTED` exige lista não vazia; mudar a lista não derruba membros ativos, mas bloqueia o método).
- Modelo híbrido recomendado: criar uma organização "pessoal" por usuário individual para permitir crescer para time.

## O que a comunidade elogia (com evidência)
- API e documentação para desenvolvedores; foco em passwordless (posts de lançamento e cobertura, HN 30375044, 29265886; pouca discussão independente).
- O CEO respondeu ponto a ponto às acusações de 2022 e descreveu PKCE em links mágicos e programa de divulgação (HN 33164969, fornecedor).

## Do que a comunidade se queixa (com evidência)
- **Segurança em 2022**: post anônimo acusou ausência de proteção contra *login CSRF* e de política de divulgação (HN 33162854, 58 pontos); comentário de funcionário de concorrente (FusionAuth) no fio; o CEO contestou parte e reconheceu limites de PKCE em webviews mobile.
- **Incerteza pós-aquisição**: "roadmap got broader and slower" (MojoAuth, concorrente) e "priorities shaped by Twilio" (WorkOS, concorrente) — ambos enviesados.
- **Sem residência de dados** (doc oficial) e MFA B2B limitado a SMS/TOTP.
- **Cobrança por conexão** semelhante a WorkOS/Frontegg (HN 44196896).

## Histórico de segurança

| ID / caso | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVEs/GHSA de SDKs | — | **Não encontrados** no OSV (`stytch`, `@stytch/vanilla-js`, `@stytch/nextjs`, `@stytch/react`, PyPI `stytch`) | — |
| Divulgação pública no HN (login CSRF, links mágicos reutilizáveis) | 2022 | Login CSRF em fluxos passwordless; ausência percebida de programa de divulgação | Vincular link/código ao dispositivo que iniciou (PKCE ou cookie de estado), uso único, e publicar VDP com safe harbor |
| Ataque de redirecionamento em links mágicos (Dfns) | 2024 | Link mágico com redirect para domínio do atacante roubava o token; Stytch passou a exigir lista de redirects pré-aprovados | Redirect URLs sempre de allowlist exata; preferir códigos a links |

## Arquitetura e decisões notáveis
- Produtos Consumer e B2B são tipos de projeto distintos, com APIs separadas (`/consumer` vs `/b2b`).
- Sessão com dois artefatos: token opaco (revogação imediata) e JWT (desempenho), à escolha do app.
- Intermediate session como máquina de estados do login (descoberta → seleção de org → MFA → sessão plena).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Modelo**: projeto (isolamento total) → Organizations (tenant do cliente, `organization_id` + `slug`) → Members (por org, `member_id` único global, e-mail único dentro da org).
- Por organização: métodos de autenticação permitidos (`allowed_auth_methods` + `RESTRICTED`), métodos de MFA permitidos, JIT por domínio e por SSO, domínios de e-mail permitidos, convites, conexões SSO/SCIM, metadados.
- RBAC: política de projeto com recursos/ações e papéis; atribuição por SSO/SCIM.
- Extensão: custom claims, metadados, trusted auth tokens, webhooks (não detalhado). Hooks no fluxo: não encontrados.
- Bancos: não se aplica. Soberania: só EUA.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não** encontrada.
- Autorização sensível ao estado / transições: **não** (RBAC recurso/ação apenas).

## Implicações para o projeto
1. **Adotar** como regra de segurança: SSO de uma organização só autentica nessa organização (ou com opt-in explícito e auditado); fazer disso teste de regressão.
2. **Adotar** estado explícito de login parcial (intermediate session) para descoberta, seleção de organização e MFA.
3. **Avaliar** no brief 4.7: membership com credenciais/MFA escopáveis por organização como opção (Stytch) vs. identidade única por tenant (Clerk/WorkOS); oferecer as duas por configuração.
4. **Adotar** regras de validação de configuração documentadas e testadas (combinações inválidas rejeitadas com erro claro; mudanças que endurecem não derrubam sessões, mas bloqueiam o método).
5. **Adotar** links mágicos vinculados ao dispositivo + redirects por allowlist exata.
6. **Oportunidade**: residência de dados e antifraude abertos (DFP é proprietário e pago por fingerprint).

## Fontes
- https://stytch.com/docs/llms.txt e https://stytch.com/docs/_llms/documentation.txt — 2026-11-03 — oficial
- https://stytch.com/docs/multi-tenant-auth/data-model.md, /build-auth/discovery.md, /enterprise-ready/org-management/org-configuration-rules.md, /authentication/passwords/cross-org.md — 2026-11-03 — oficial
- https://stytch.com/docs/multi-tenant-auth/manage-sessions/intermediate-sessions/overview.md, /manage-sessions/jwts-and-tokens.md, /authentication/mfa/overview.md, /authentication/sso/multi-organization.md, /authentication/user-impersonation.md, /enterprise-ready/admin-portal.md, /enterprise-ready/rbac/create-rbac-policy.md — 2026-11-03 — oficial
- https://stytch.com/docs/fraud-risk/device-fingerprinting/overview.md, /get-started/guides/ai-agents-and-apps.md — 2026-11-03 — oficial
- https://stytch.com/docs/resources/policies/security-and-trust/compliance/data-compliance.md — 2026-11-03 — oficial
- https://stytch.com/pricing — 2026-11-03 — oficial
- https://stytch.com/docs/resources/changelog (via busca: aquisição concluída) — 2026-11-03 — oficial
- https://news.ycombinator.com/item?id=33162854, 33164969, 39470158, 45765090, 29274791, 44196896 — 2026-11-03 — comunidade (33164969 e 39470158 são do fornecedor)
- https://www.dfns.co/article/the-magic-link-vulnerability — 2026-11-03 — terceiro (fornecedor de carteiras)
- https://mojoauth.com/blog/mojoauth-vs-stytch-passwordless-api e https://workos.com/blog/stytch-alternatives (via busca) — 2026-11-03 — concorrentes
- https://api.osv.dev/v1/query (pacotes listados) — 2026-11-03 — base pública
- https://api.npmjs.org/downloads/point/last-month/stytch — 2026-11-03 — npm
