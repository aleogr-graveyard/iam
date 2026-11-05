# Frontegg

| Campo | Valor |
|---|---|
| Categoria | CIAM/IDaaS B2B "user management" com **portal de autoatendimento** embutível, entitlements (planos, feature flags, RBAC) e, desde 2025–26, a linha **Agen** (governança de agentes de IA: Agen for SaaS / Agen for Work) |
| Site / repositório | https://frontegg.com · docs https://developers.frontegg.com · SDKs em https://github.com/frontegg (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário; SDKs abertos (licença não verificada). "Pay as you go" grátis até 7.500 usuários, organizações ilimitadas, 5 conexões enterprise (SSO/SCIM); Enterprise sob consulta (página oficial). Preço público além da franquia: **não encontrado** (agregadores divergem; em 2024 migrou para freemium, segundo PricingSaaS) |
| Stack | Fechada; não verificada |
| Maturidade | Empresa israelense (conhecimento prévio: fundada ~2019). `@frontegg/react` ~570 mil e `@frontegg/client` ~540 mil downloads/mês; `@frontegg/nextjs` só ~27 mil (npm, out/2026). G2 ~4,8 com ~360 avaliações (via busca) |
| Acesso | 2026-11-03 |

> Limitações: GitHub e Reddit bloqueados; página de regiões (docs.frontegg.com) só lida via trecho de busca; comunidade independente quase ausente no HN.

## Resumo
- Foco em entregar "enterprise-readiness" pronta: login box, **Admin Portal** (o cliente gerencia usuários, SSO, SCIM, MFA, sessões, IP, API tokens, auditoria) e Security Center.
- Modelo B2B: **accounts (tenants)** com usuários em vários tenants, **hierarquia de sub-accounts** (inclusive revenda), login customizado por account, aplicações múltiplas atribuídas por account.
- Políticas por tenant amplas: MFA, sessão, senha, restrição de IP e de domínio, papéis e tokens por tenant.
- Orquestração: **Flows** (no-code, com sub-flows, webhooks, prehooks e geração por linguagem natural).
- Linha Agen inclui **fluxos de aprovação com mínimo de aprovadores por etapa** para chamadas de ferramentas de agentes.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | Senha, link mágico/OTP, passkeys, SMS, social, SSO SAML/OIDC por account (autoatendimento para Okta, Entra ID, Google, JumpCloud, OneLogin, Ping, Rippling). LDAP e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ● | TOTP, SMS, passkeys, **WhatsApp via prehook `SEND_SMS`** (entrega por provedor do cliente), "lembrar dispositivo" por aplicação, **step-up** com SDKs, MFA condicional em Flows. Push: não encontrado |
| 3 | Ciclo de vida da conta | ● | Convites, SCIM, JIT por SSO, bloqueio de account (tenant), usuários inativos (*stale users*), exclusão. Verificação de idade: não encontrada |
| 4 | Sessões e tokens | ◐ | Duração, máximo de sessões simultâneas, revogação individual e total, políticas por tenant, templates/claims de token. DPoP, back-channel logout, rotação de refresh: não verificados |
| 5 | Protocolos | ◐ | OIDC/OAuth (hosted login), SAML/OIDC como SP, SCIM, token exchange (M2M). Frontegg como IdP SAML para terceiros, CIBA, PAR, FAPI, device flow: não encontrados |
| 6 | Máquinas e agentes | ◐ | M2M com *client credentials* e *access tokens* em contexto de usuário ou de tenant, autoatendimento de tokens pelo cliente, token exchange; **Agen**: gateway MCP, políticas por ferramenta, autorização de agentes |
| 7 | Detecção e defesa | ● | Security rules: bot detection, senha vazada, força bruta, restrição por país, fingerprint de dispositivo, credibilidade de e-mail, viagem impossível, IP suspeito; **overrides** por usuário/tenant/IP com prioridade; painel de eventos mitigados/não mitigados |
| 8 | Privacidade e conformidade | ◐ | Conta hospedada **em uma região: UE (Irlanda) ou EUA (us-east-1)**, conforme o portal usado (doc oficial via busca; pode estar desatualizada). GDPR. Outras certificações: não verificadas |
| 9 | Experiência e administração | ● | Admin Portal embutido ou hospedado (módulos pessoais e de workspace), login box por tenant, Flows, audit logs, webhooks/integrações, impersonação, entitlements (planos, feature flags, RBAC) |
| 10 | Operação e migração | ◐ | Ambientes; até 200 aplicações por ambiente; só SaaS; ferramentas de migração não detalhadas |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Admin Portal** mais completo do grupo: o cliente final administra usuários, grupos, papéis, SSO, SCIM, MFA, sessões, restrições de IP/domínio, API tokens e auditoria do próprio tenant.
- **Hierarquia de accounts** (pai → filhos, vários níveis) com permissões específicas (`fe.account-hierarchy.read.subAccount`) para cenários de revenda.
- **Login por account** (subdomínio ou parâmetro): UI e métodos isolados por tenant; não permite `switchTenant` entre accounts com login próprio.
- **Entitlements** = RBAC + planos + feature flags avaliados juntos, com agente local para avaliação (avaliado em AuthZ).
- **Overrides de regras de segurança** por usuário, tenant ou IP sem alterar a regra global.
- **Approval flows** (Agen): cadeia sequencial de etapas, cada uma com grupos/e-mails e **"required minimum approvers"**, lembretes e canais (e-mail, SMS).

## O que a comunidade elogia (com evidência)
- Suporte rápido (inclusive via Slack) e Admin Portal fácil de usar (avaliações G2, via busca; avaliações possivelmente incentivadas).
- Recomendação espontânea em HN como "has everything you're looking for" (HN 36103602, autor não identificado como funcionário; outro comentário no mesmo período é de funcionário, HN 35632352).

## Do que a comunidade se queixa (com evidência)
- **APIs não documentadas** e erros pouco claros; faltam exemplos de cenários complexos (avaliações G2, via busca).
- **Pouca customização visual** do login box e do Admin Portal; pedido para adicionar seções próprias ao portal hospedado (G2, Capterra, via busca).
- **Cobrança por conexão** SSO/SCIM semelhante a WorkOS/Stytch (HN 44196896).
- Trustpilot 2,8 (poucas avaliações; conteúdo não verificado).
- Uso baixo do SDK Next.js comparado a Clerk/WorkOS (npm), sinal de menor adoção no ecossistema JS moderno (inferência).

## Histórico de segurança
- CVEs/advisories: **não encontrados** no OSV para `@frontegg/nextjs`, `@frontegg/react`, `@frontegg/client`, `@frontegg/rest-api`, `@frontegg/js`; busca web sem incidentes.
- Lição de design: overrides de regras de segurança são poderosos e perigosos; precisam de auditoria, prioridade explícita e alerta (premissa "alertar em vez de bloquear").

## Arquitetura e decisões notáveis
- Vendor → ambientes → aplicações (até 200) → accounts (tenants) → sub-accounts; aplicações atribuídas a accounts e usuários.
- Prehooks (síncronos) e webhooks (assíncronos) como pontos de extensão; WhatsApp implementado via prehook de SMS.
- Flows com IA que gera o fluxo a partir de descrição, com revisão humana.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Modelo**: usuário pode pertencer a vários tenants e trocar de tenant (`switchTenant`); papéis por tenant; tokens M2M de tenant.
- Por tenant: SSO (vários provedores), SCIM, MFA, política de sessão, senha, IP/domínio permitidos, papéis, login customizado, planos/entitlements, aplicações atribuídas, bloqueio do tenant.
- Hierarquia: sub-accounts com usuários no pai e nos filhos.
- Isolamento total: ambiente/conta separados; login por account dá isolamento de UX e de métodos, não de identidade.
- Bancos: não se aplica. Soberania: UE ou EUA por conta.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **parcial**. Agen for Work/SaaS tem fluxos de aprovação em cadeia com "Required minimum approvers" por etapa, mas só para chamadas de ferramentas de agentes de IA (política "Request approval"), não para ações arbitrárias de usuários ou recursos da aplicação (docs Approval flows).
- Autorização sensível ao estado / transições: **não** encontrada (entitlements = papéis + planos + flags).

## Implicações para o projeto
1. **Adotar** um portal de autoatendimento por organização cobrindo SSO, SCIM, MFA, sessões, IP/domínio, tokens de API e auditoria; é o que mais se elogia no Frontegg.
2. **Adotar** hierarquia de organizações (revenda, unidades) com permissões próprias para gerir sub-organizações.
3. **Adotar** overrides de regras de defesa por usuário/organização/IP, sempre auditados e com alerta.
4. **Considerar** no AuthZ o modelo de aprovação em etapas com quórum mínimo por etapa (Agen) como referência de UX; o diferencial do projeto seria generalizá-lo para qualquer ação/transição, não só agentes.
5. **Evitar** canais (WhatsApp) só via gambiarra de hook; oferecer canal de primeira classe com templates aprovados (brief 5.1).
6. **Oportunidade**: documentação de API completa e gerada da especificação (queixa recorrente).

## Fontes
- https://developers.frontegg.com/llms.txt — 2026-11-03 — oficial
- https://developers.frontegg.com/ciam/guides/management/manage-accounts/overview.md, /hierarchies.md, /login-per-account.md, /multi-apps/overview.md — 2026-11-03 — oficial
- https://developers.frontegg.com/ciam/guides/security-center/overview.md, /security-rules/overrides.md, /session-management/overview.md — 2026-11-03 — oficial
- https://developers.frontegg.com/ciam/guides/use-cases/whatsapp-mfa.md, /authentication/m2m/overview.md, /step-up/intro.md, /integrations/flows/overview.md, /admin-portal/intro.md, /authorization/entitlements-rbac-overview.md — 2026-11-03 — oficial
- https://developers.frontegg.com/agen-for-work/approval-flows/overview.md e /agen-for-saas/policies/approval-flows.md — 2026-11-03 — oficial
- https://frontegg.com/pricing — 2026-11-03 — oficial
- https://docs.frontegg.com/docs/what-countries-and-regions-is-frontegg-deployed-on (trecho via WebSearch) — 2026-11-03 — oficial, possivelmente desatualizado
- https://pricingsaas.com/companies/frontegg/diffs/2024Q2 (via busca) — 2026-11-03 — terceiro
- G2/Capterra/Trustpilot (via WebSearch) — 2026-11-03 — sites de avaliação, possível incentivo
- https://news.ycombinator.com/item?id=44196896, 36103602, 35632352 — 2026-11-03 — comunidade (35632352 é funcionário)
- https://api.osv.dev/v1/query (pacotes listados) — 2026-11-03 — base pública
- https://api.npmjs.org/downloads/point/last-month/@frontegg/react — 2026-11-03 — npm
